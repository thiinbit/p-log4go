// Copyright 2020 @thiinbit(thiinbit@gmail.com). All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package plog4go

import "log"

// file config
const (
	// DefaultLogRootPath log root dir
	DefaultLogRootPath string = "./run-log"

	// Default default log file name
	DefaultLogFile string = "default.log"

	// default flag
	DefaultLogFlag int = log.Ldate|log.Ltime
)

