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

package base

type UseScope int

type Location int

const (
	Global UseScope = iota
	Project
	Session
)

const (
	OriginalLocation Location = iota
	GlobalLocation
	ShellLocation
)

func (s UseScope) String() string {
	switch s {
	case Global:
		return "global"
	case Project:
		return "project"
	case Session:
		return "session"
	default:
		return "unknown"
	}
}

func (s Location) String() string {
	switch s {
	case GlobalLocation:
		return "global"
	case ShellLocation:
		return "shell"
	case OriginalLocation:
		return "original"
	default:
		return "unknown"
	}
}
