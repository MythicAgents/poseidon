//go:build shared
// +build shared

// This code is borrowed and slightly modified from https://github.com/MythicAgents/merlin/blob/efde48c42ed6dc364258698ef3a49009c684dd9f/Payload_Type/merlin/agent/shared.go
// Merlin is a post-exploitation command and control framework.
// This file is part of Merlin.
// Copyright (C) 2023 Russel Van Tuyl

// Merlin is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.

// Merlin is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Merlin.  If not, see <http://www.gnu.org/licenses/>.

package main

/*
 #include "poseidon.h"
*/
import "C"
