// Copyright (C) 2021 Dakota Walsh
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package gemtext provides a gemtext Renderer for use with the goldmark
// library: https://github.com/yuin/goldmark
//
// Gemtext is a lightweight markup language for use over the Gemini protocol,
// it's more or less a subset of Markdown so by definition some source material
// MUST be lost during a proper conversion. This library offers several
// configuration options for different ways of handling this simplification.
// You can learn more about Genini here: https://gemini.circumlunar.space/
package gemtext
