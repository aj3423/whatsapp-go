package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"str"

	"github.com/fatih/color"
)

func main() {
	mm_path, _ := os.Getwd()

	target_proto_dir := mm_path + "/pb"

	// 1. clean up files not *.pb.cc|*.pb.h
	color.HiYellow("cleanup %s ...", target_proto_dir)

	var files []string

	err := filepath.Walk(target_proto_dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Ext(path) == ".proto" {
				// save to files
				files = append(files, path)
			} else if str.ContainsAny(path, ".pb.h", ".pb.cc", ".pb.go") {
				os.Remove(path)
				return nil
			}
			// fmt.Println(path, info.Size())
			return nil
		})
	if err != nil {
		panic(err)
	}

	//fmt.Println(files)
	//os.Exit(1)

	// 2. generate
	color.HiYellow("compiling .proto in %s ...", target_proto_dir)

	protoc := `protoc`

	for _, f := range files {

		var cmds []string
		// append this file path
		cmds = append(cmds, fmt.Sprintf(`--proto_path=%s`, filepath.Dir(f)))

		cmds = append(cmds, fmt.Sprintf(`--proto_path=%s`, mm_path+"/pb"))
		cmds = append(cmds, fmt.Sprintf(`--go_out=%s`, filepath.Dir(f)))
		cmds = append(cmds, fmt.Sprintf(`%s`, f))

		runner := exec.Command(protoc, cmds...)
		runner.Stdout = os.Stdout
		runner.Stderr = os.Stderr
		err := runner.Run()
		if err != nil {
			color.HiRed(err.Error())
			os.Exit(1)
		}
	}

	color.HiGreen("Done")
}
