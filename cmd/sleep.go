package cmd

import (
	"fmt"
	"time"
)

// like /bin/sleep without options
//
//	sleep NUMBER[SUFFIX]...
var Sleep = &SimpleMatchCommand{
	Name: "sleep",
	F: func(args []string, e *Env) int {
		duration := time.Duration(0)

		for _, arg := range args[1:] {
			var val float64
			if _, err := fmt.Sscanf(arg, "%fs", &val); err == nil {
				duration += time.Duration(val) * time.Second
			} else if _, err := fmt.Sscanf(arg, "%fm", &val); err == nil {
				duration += time.Duration(val) * time.Minute
			} else if _, err := fmt.Sscanf(arg, "%fh", &val); err == nil {
				duration += time.Duration(val) * time.Hour
			} else if _, err := fmt.Sscanf(arg, "%fd", &val); err == nil {
				duration += time.Duration(val) * time.Hour * 24
			} else {
				e.Error(fmt.Errorf("invalid time interval '%s'", arg))

				return 1
			}
		}

		time.Sleep(duration)

		return 0
	},
}
