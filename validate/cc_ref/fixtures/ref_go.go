package fixtures

import "os"

func select_statement(ch1, ch2 chan int) int {
	select {
	case v := <-ch1:
		return v
	case v := <-ch2:
		return v
	default:
		return 0
	}
}

func type_switch(v interface{}) string {
	switch v.(type) {
	case int:
		return "int"
	case string:
		return "string"
	default:
		return "unknown"
	}
}

func defer_statement(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func go_statement() {
	done := make(chan bool)
	go func() {
		done <- true
	}()
	<-done
}
