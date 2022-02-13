package kvdb

import (
	"fmt"
	"io"
	"strings"

	"github.com/fsouza/stream"
)

type DB struct {
	data   map[string]string
	writer io.Writer
	cmds   map[string]func([]string)
}

func NewDB(w io.Writer) *DB {
	db := &DB{
		data:   map[string]string{},
		writer: w,
	}
	db.cmds = map[string]func([]string){
		"set": db.set,
		"get": db.get,
		"del": db.del,
	}
	return db
}

func (db *DB) Execute(cmd string, args []string) {
	cmdFn := db.cmds[cmd]
	cmdFn(args)
}

func (db *DB) set(args []string) {
	db.data[args[0]] = args[1]
}

func (db *DB) get(args []string) {
	if v, ok := db.data[args[0]]; ok {
		fmt.Fprintln(db.writer, v)
	}
}

func (db *DB) del(args []string) {
	delete(db.data, args[0])
}

func ProcessCommands(input io.Reader, output io.Writer) {
	s := stream.FromReader(input)
	stream.Fold(s, NewDB(output), func(db *DB, line string) *DB {
		cmd := strings.Fields(line)
		db.Execute(cmd[0], cmd[1:])
		return db
	})
}
