# Nullable
Golang nullable package powered by generics

## Install

```shell
go get github.com/rlshukhov/nullable
```

## Example

```go
package main

import (
	"fmt"
	"github.com/rlshukhov/nullable"
	"gopkg.in/yaml.v3"
)

type name struct {
	Name nullable.Nullable[string] `yaml:"name"`
	Age  nullable.Nullable[uint64] `yaml:"age"`
}

func main() {
	var nilValue *string
	value := "Alice"

	data := []name{
		{nullable.FromValue("Bob"), nullable.Null[uint64]()},
		{nullable.FromValue(""), nullable.FromValue[uint64](21)},
		{nullable.FromPointer(nilValue), nullable.FromValue[uint64](34)},
		{nullable.FromPointer(&value), nullable.FromValue[uint64](45)},
	}

	out, err := yaml.Marshal(&data)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

	var parsedData []name
	err = yaml.Unmarshal(out, &parsedData)

	for _, d := range parsedData {
		fmt.Printf("Name:\t%s\tIsNull:\t%t\tHasValue:\t%t\n", d.Name.GetValue(), d.Name.IsNull(), d.Name.HasValue())
		fmt.Printf("Age:\t%d\tIsNull:\t%t\tHasValue:\t%t\n\n", d.Age.GetValue(), d.Age.IsNull(), d.Age.HasValue())
	}
}
```

```shell
rlshukhov@MacBook-Pro-Lane main % go run main.go
- name: Bob
  age: null
- name: ""
  age: 21
- name: null
  age: 34
- name: Alice
  age: 45

Name:   Bob     IsNull: false   HasValue:       true
Age:    0       IsNull: true    HasValue:       false

Name:           IsNull: false   HasValue:       true
Age:    21      IsNull: false   HasValue:       true

Name:           IsNull: true    HasValue:       false
Age:    34      IsNull: false   HasValue:       true

Name:   Alice   IsNull: false   HasValue:       true
Age:    45      IsNull: false   HasValue:       true
```
