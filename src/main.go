package main

func main() {
	// Recover from a panic from https://github.com/golang/go/wiki/PanicAndRecover

	/*defer func() {
		if r := recover(); r != nil {
			var ok bool
			var err error
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()*/

}

func run() /*error*/ {
	/*
		err := something()
		if err != nil {
			return err
		}
		// etc
	*/

}

func RecoverIfPanic() {

}
