package iotmaker_capibaribe_module

import (
	"fmt"
)

func ExampleAnalytics_OnExecutionStart() {

	a := Analytics{}

	// before execution start
	fmt.Printf("NumberCurrentExecutions: %v\n", a.NumberCurrentExecutions)
	fmt.Printf("timeOnStarEvent equals to 0: %v\n", a.timeOnStarEvent.UnixNano())
	a.OnExecutionStart()

	fmt.Printf("NumberCurrentExecutions: %v\n", a.NumberCurrentExecutions)
	fmt.Printf("timeOnStarEvent greater than 0: %v\n", a.timeOnStarEvent.UnixNano() > 0)

	// Output:
	// NumberCurrentExecutions: 0
	// timeOnStarEvent equals to 0: -6795364578871345152
	// NumberCurrentExecutions: 1
	// timeOnStarEvent greater than 0: true
}
