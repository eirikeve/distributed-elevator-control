package elevorderevaluation

import(
	"testing"
)

/*
* Test if finding maximum index, i.e. element with largest integer value, is working properly
*/
func TestMaxIndex(t *testing.T){
	durations := []int{}
	index := findMaxIndex(durations)
	println("Max index:",index)
	println("Containg the value", durations[index])

}

