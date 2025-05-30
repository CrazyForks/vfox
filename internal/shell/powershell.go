/*
 *    Copyright 2025 Han Li and contributors
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package shell

import (
	"fmt"
	"regexp"

	"github.com/version-fox/vfox/internal/env"
)

// Based on https://github.com/direnv/direnv/blob/master/internal/cmd/shell_pwsh.go
type pwsh struct{}

// Pwsh shell instance
var Pwsh Shell = pwsh{}

const hook = `

<#
All environment variables must be set in global scope
DO NOT PUT IN MODULE.
#>
{{.EnvContent}}

$env:__VFOX_PID = $pid;

# remove any existing dynamic module of vfox
if ($null -ne (Get-Module -Name "version-fox")) {
    Remove-Module -Name "version-fox" -Force
}

# create a new module to override the prompt function
New-Module -Name "version-fox" -ScriptBlock {

    <#
    Due to a bug in PowerShell, we have to cleanup first when the shell open.
    #>
    & '{{.SelfPath}}' env --cleanup 2>$null | Out-Null;

    $originalPrompt = $function:prompt;
    $OutputEncoding = [console]::InputEncoding = [console]::OutputEncoding = [Text.UTF8Encoding]::UTF8;

    $promptFunction = {
        $export = &"{{.SelfPath}}" env -s pwsh;
        if ($export) {
            Invoke-Expression -Command $export;
        }
        &$originalPrompt;
    }
    $function:prompt = $promptFunction

    <#
     There is a bug here.
     When powershell is closed via the x button, this event will not be fired.
     See https://github.com/PowerShell/PowerShell/issues/8000
    #>
    $subscription = Register-EngineEvent -SourceIdentifier PowerShell.Exiting -Action {
        &"{{.SelfPath}}" env --cleanup;
    }

    # perform cleanup on removal so a new initialization in current session works
    $ExecutionContext.SessionState.Module.OnRemove += {
        $function:prompt = $originalPrompt
        Unregister-Event -SubscriptionId $subscription.Id
    }
} | Import-Module -Global
`

func (sh pwsh) Activate(config ActivateConfig) (string, error) {
	return hook, nil
}

func (sh pwsh) Export(e env.Vars) (out string) {
	for key, value := range e {
		if value == nil {
			out += sh.unset(key)
		} else {
			out += sh.export(key, *value)
		}
	}
	return out
}

func (sh pwsh) export(key, value string) string {
	value = sh.escape(value)
	if !regexp.MustCompile(`'.*'`).MatchString(value) {
		value = fmt.Sprintf("'%s'", value)
	}
	return fmt.Sprintf("$env:%s=%s;", sh.escape(key), value)
}

func (sh pwsh) unset(key string) string {
	return fmt.Sprintf("Remove-Item -Path 'env:/%s';", sh.escape(key))
}

func (pwsh) escape(str string) string {
	return PowerShellEscape(str)
}

func PowerShellEscape(str string) string {
	if str == "" {
		return "''"
	}
	in := []byte(str)
	out := ""
	i := 0
	l := len(in)
	escape := false

	hex := func(char byte) {
		escape = true
		out += fmt.Sprintf("\\x%02x", char)
	}

	backslash := func(char byte) {
		escape = true
		out += string([]byte{BACKTICK, char})
	}

	escaped := func(str string) {
		escape = true
		out += str
	}

	quoted := func(char byte) {
		escape = true
		out += string([]byte{char})
	}

	literal := func(char byte) {
		out += string([]byte{char})
	}

	for i < l {
		char := in[i]
		switch {
		case char == ACK:
			hex(char)
		case char == TAB:
			escaped("`t")
		case char == LF:
			escaped("`n")
		case char == CR:
			escaped("`r")
		case char <= US:
			hex(char)
		// case char <= AMPERSTAND:
		// 	quoted(char)
		case char == SINGLE_QUOTE:
			backslash(char)
		case char <= PLUS:
			quoted(char)
		case char <= NINE:
			literal(char)
		// case char <= QUESTION:
		// 	quoted(char)
		case char <= UPPERCASE_Z:
			literal(char)
		// case char == OPEN_BRACKET:
		// 	quoted(char)
		// case char == BACKSLASH:
		// 	quoted(char)
		case char == UNDERSCORE:
			literal(char)
		// case char <= CLOSE_BRACKET:
		// 	quoted(char)
		// case char <= BACKTICK:
		// 	quoted(char)
		// case char <= TILDA:
		// 	quoted(char)
		case char == DEL:
			hex(char)
		default:
			quoted(char)
		}
		i++
	}

	if escape {
		out = "'" + out + "'"
	}

	return out
}
