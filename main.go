package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// ./main.go -operation «add» -item ‘{«id»: "1", «email»: «email@test.com», «age»: 23}’ -fileName «users.json» -operation
func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

// id, item, operation and fileName
type Arguments map[string]string

type Item struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// 1. add list - storing in the file
// 2. find by id
// 3. remove

func Perform(args Arguments, writer io.Writer) error {
	op, file, item, id := args["operation"], args["fileName"], args["item"], args["id"]
	if op == "" {
		return errors.New("-operation flag has to be specified")
	}

	if file == "" {
		return errors.New("-fileName flag has to be specified")
	}
	found := false
	for _, v := range []string{"add", "list", "findById", "remove"} {
		if v == op {
			found = true
		}
	}

	if !found {
		return errors.New("Operation abcd not allowed!")
	}

	if op == "add" && item == "" {
		return errors.New("-item flag has to be specified")
	}

	if (op == "findById" || op == "remove") && id == "" {
		return errors.New("-id flag has to be specified")
	}

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0744)
	if err != nil {
		return err
	}

	defer f.Close()

	body, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	switch op {
	case "add":
		It, Its := Item{}, []Item{}
		if err := json.Unmarshal([]byte(item), &It); err != nil {
			return err
		}

		if len(body) != 0 {
			if err := json.Unmarshal(body, &Its); err != nil {
				return err
			}

			for _, it := range Its {
				if it.Id == It.Id {
					writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", It.Id)))
				}
			}
		}

		Its = append(Its, It)
		res, err := json.Marshal(Its)
		if err != nil {
			return err
		}

		if _, err := f.Write(res); err != nil {
			return err
		}

	case "list":
		if _, err := writer.Write(body); err != nil {
			return err
		}
	case "findById":
		var isFound bool
		var items []Item
		if err = json.Unmarshal(body, &items); err != nil {
			return err
		}
		for _, item := range items {
			if item.Id == args["id"] {
				isFound = true
				res, err := json.Marshal(item)
				if err != nil {
					return err
				}
				if _, err = writer.Write(res); err != nil {
					return err
				}
			}
		}
		if !isFound {
			if _, err = writer.Write([]byte("")); err != nil {
				return err
			}
		}
	case "remove":
		var items []Item
		if err = json.Unmarshal(body, &items); err != nil {
			return err
		}
		var arr []Item
		for i, v := range items {
			if v.Id == args["id"] {
				arr = append(items[:i], items[(i+1):]...)
				break
			}
		}
		if len(arr) == 0 {
			writer.Write([]byte(fmt.Sprintf("Item with id %s not found", args["id"])))
			return nil
		}
		res, err := json.Marshal(arr)
		if err != nil {
			return err
		}
		f.Truncate(0)

		if _, err = f.WriteAt(res, 0); err != nil {
			return err
		}
	}

	return nil
}

func parseArgs() (args Arguments) {
	args = make(Arguments, 4)
	op := flag.String("operation", "none", "-operation flag has to be specified")
	file := flag.String("fileName", "none", "-fileName flag has to be specified")

	flag.Parse()
	args["operation"] = *op
	args["fileName"] = *file
	return
}
