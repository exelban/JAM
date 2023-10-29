package main

func main() {
	hold := make(chan interface{})

	for i := 0; i < 2; i++ {
		go func() {
			for {
			}
		}()
	}

	<-hold
}
