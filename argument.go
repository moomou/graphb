package graphb

import (
	"fmt"
	"regexp"
)

type argumentValue interface {
	stringChan() <-chan string
}

type Argument struct {
	Name   string
	Value  argumentValue
	IsFunc bool
}

func (a *Argument) stringChan() <-chan string {
	tokenChan := make(chan string)
	go func() {
		tokenChan <- a.Name
		if a.IsFunc {
			tokenChan <- tokenLP
		} else {
			tokenChan <- ":"
		}
		for str := range a.Value.stringChan() {
			if a.IsFunc && (str == tokenLB || str == tokenRB) {
				continue
			} else if a.IsFunc && str == tokenColumn {
				if a.Name == "has" {
					break
				} else {
					str = tokenComma
				}
			}
			tokenChan <- str
		}
		if a.IsFunc {
			tokenChan <- tokenRP
		}
		close(tokenChan)
	}()
	return tokenChan
}

func ArgumentAny(name string, value interface{}) (Argument, error) {
	switch v := value.(type) {
	case bool:
		return ArgumentBool(name, v), nil
	case []bool:
		return ArgumentBoolSlice(name, v...), nil

	case int:
		return ArgumentInt(name, v), nil
	case []int:
		return ArgumentIntSlice(name, v...), nil

	case string:
		return ArgumentString(name, v), nil
	case []string:
		return ArgumentStringSlice(name, v...), nil

	case *regexp.Regexp:
		return ArgumentRegex(name, v), nil

	default:
		return Argument{}, ArgumentTypeNotSupportedErr{Value: value}
	}
}

func ArgumentBool(name string, value bool) Argument {
	return Argument{name, argBool(value), false}
}

func ArgumentInt(name string, value int) Argument {
	return Argument{name, argInt(value), false}
}

func ArgumentString(name string, value string) Argument {
	return Argument{name, argString(value), false}
}

func ArgumentRegex(name string, value *regexp.Regexp) Argument {
	va := argRegexp(value.String())
	return Argument{name, va, false}
}

func ArgumentBoolSlice(name string, values ...bool) Argument {
	return Argument{name, argBoolSlice(values), false}
}

func ArgumentIntSlice(name string, values ...int) Argument {
	return Argument{name, argIntSlice(values), false}
}

func ArgumentStringSlice(name string, values ...string) Argument {
	return Argument{name, argStringSlice(values), false}
}

// ArgumentCustomType returns a custom GraphQL type's argument representation, which could be a recursive structure.
func ArgumentCustomType(name string, values ...Argument) Argument {
	return Argument{name, argumentSlice(values), false}
}

func ArgumentFuncType(name string, values ...Argument) Argument {
	return Argument{name, argumentSlice(values), true}
}

/////////////////////////////
// Primitive Wrapper Types //
/////////////////////////////

// argBool represents a boolean value.
type argBool bool

func (v argBool) stringChan() <-chan string {
	tokenChan := make(chan string)
	go func() {
		tokenChan <- fmt.Sprintf("%t", v)
		close(tokenChan)
	}()
	return tokenChan
}

// argInt represents an integer value.
type argInt int

func (v argInt) stringChan() <-chan string {
	tokenChan := make(chan string)
	go func() {
		tokenChan <- fmt.Sprintf("%d", v)
		close(tokenChan)
	}()
	return tokenChan
}

// argString represents a string value.
type argString string

func (v argString) stringChan() <-chan string {
	tokenChan := make(chan string)
	go func() {
		tokenChan <- fmt.Sprintf(`"%s"`, v)
		close(tokenChan)
	}()
	return tokenChan
}

//////////////////////////////////
// Primitive List Wrapper Types //
//////////////////////////////////

// argBoolSlice implements valueSlice
type argBoolSlice []bool

func (s argBoolSlice) stringChan() <-chan string {
	tokenChan := make(chan string)
	go func() {
		tokenChan <- "["
		for i, v := range s {
			if i != 0 {
				tokenChan <- ","
			}
			tokenChan <- fmt.Sprintf("%t", v)
		}
		tokenChan <- "]"
		close(tokenChan)
	}()
	return tokenChan
}

// argIntSlice implements valueSlice
type argIntSlice []int

func (s argIntSlice) stringChan() <-chan string {
	tokenChan := make(chan string)
	go func() {
		tokenChan <- "["
		for i, v := range s {
			if i != 0 {
				tokenChan <- ","
			}
			tokenChan <- fmt.Sprintf("%d", v)
		}
		tokenChan <- "]"
		close(tokenChan)
	}()
	return tokenChan
}

// argStringSlice implements valueSlice
type argStringSlice []string

func (s argStringSlice) stringChan() <-chan string {
	tokenChan := make(chan string)
	go func() {
		tokenChan <- "["
		for i, v := range s {
			if i != 0 {
				tokenChan <- ","
			}
			tokenChan <- fmt.Sprintf(`"%s"`, v)
		}
		tokenChan <- "]"
		close(tokenChan)
	}()
	return tokenChan
}

type argumentSlice []Argument

func (s argumentSlice) stringChan() <-chan string {
	tokenChan := make(chan string)

	go func() {
		tokenChan <- tokenLB
		for i, v := range s {
			if i != 0 {
				tokenChan <- tokenComma
			}
			for str := range v.stringChan() {
				tokenChan <- str
			}
		}
		tokenChan <- tokenRB
		close(tokenChan)
	}()
	return tokenChan
}

type argRegexp string

func (re argRegexp) stringChan() <-chan string {
	tokenChan := make(chan string)

	go func() {
		tokenChan <- fmt.Sprintf("%s", re)
		close(tokenChan)
	}()
	return tokenChan
}
