package placeholder

import (
	"fmt"
	"regexp"
	"strings"
)

var PhRe = regexp.MustCompile(`\${{[ ]{1}([^{}\s]+)[ ]{1}}}`)

const Left = "${{ "
const Right = " }}"

type Type string

func (t Type) String() string {
	return string(t)
}

const (
	ContextType Type = "contexts"
	InputType   Type = "inputs"
	OutputType  Type = "outputs"
	RandomType  Type = "randoms"
)

type Handler func(placeholder string, values ...string) error

func MatchHolderFromHandler(needMatchString string, handlers map[Type]Handler) error {
	matchStrings := PhRe.FindAllString(needMatchString, -1)
	for _, placeholder := range matchStrings {

		value := strings.TrimLeft(placeholder, Left)
		value = strings.TrimRight(value, Right)

		ss := strings.SplitN(value, ".", 2)

		split := strings.Split(value, ".")

		switch ss[0] {
		case ContextType.String():
			if handlers[ContextType] == nil {
				continue
			}
			if len(split) != 2 {
				return fmt.Errorf("%v placeholder %v Format problem, use ${{ %v.xxx }}", ContextType, placeholder, ContextType)
			}
			err := handlers[ContextType](placeholder, split[0], split[1])
			if err != nil {
				return err
			}
		case InputType.String():
			if handlers[InputType] == nil {
				continue
			}
			if len(split) != 2 {
				return fmt.Errorf("%v placeholder %v Format problem, use ${{ %v.xxx }}", InputType, placeholder, InputType)
			}

			err := handlers[InputType](placeholder, split[0], split[1])
			if err != nil {
				return err
			}
		case OutputType.String():
			if handlers[OutputType] == nil {
				continue
			}
			if len(split) != 3 {
				return fmt.Errorf("%v placeholder %v Format problem, use ${{ %v.alias.xxx }}", OutputType, placeholder, OutputType)
			}
			err := handlers[OutputType](placeholder, split[0], split[1], split[2])
			if err != nil {
				return err
			}
		case RandomType.String():
			if handlers[RandomType] == nil {
				continue
			}
			if len(split) != 3 {
				return fmt.Errorf("%v placeholder %v Format problem, use ${{ %v.alias.xxx }}", RandomType, placeholder, RandomType)
			}
			err := handlers[RandomType](placeholder, split[0], split[1])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
