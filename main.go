package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	arguments := os.Args
	if len(arguments) < 5 {
		fmt.Println("Usage [path] [command] [arg 1] [arg 2]")
		return
	}

	path := arguments[1]
	command := arguments[2]

	arg1 := arguments[3]
	arg2 := arguments[4]

	log.Printf("[DEBUG]: path: %s, command: %s \n", path, command)

	path, err := filepath.Abs(path)
	if err != nil {
		log.Println("File dir is wrong.", err)

		return
	}

	log.Printf("[DEBUG]: path absolute: %s\n", path)

	info, err := os.Stat(path)
	if err != nil {
		log.Println("Provided path is not found.", err)

		return
	}
	if !info.IsDir() {
		log.Println("Provided path is not directory")

		return
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println("Error reading provided directory.", err)

		return
	}
	if len(files) == 0 {
		log.Println("The directory is empty")

		return
	}

	for _, file := range files {
		key := file.Name()

		log.Printf("[DEBUG] checking filename %s containing '='.", key)

		if strings.Contains(key, "=") {

			log.Printf("The filename %s contains prohibited char '='", key)

			continue
		}

		filename := filepath.Join(path, key)

		log.Printf("[DEBUG] file in path: %s\n", filename)

		info, err := os.Stat(filename)

		if err != nil {
			log.Printf("Error getting info about file %s in provided"+
				" directory. %s", filename, err)
			continue
		}

		log.Printf("[DEBUG] Checking if file  %s, is directory\n", filename)

		if info.IsDir() {
			log.Printf("The file %s is directory\n", filename)
			continue
		}

		log.Printf("[DEBUG] Checking if file  %s, is empty (size = 0)\n",
			filename)

		if info.Size() == 0 {
			log.Printf("The file %s is empty we need to delete env variable"+
				" %s\n",
				filename, key)
			if os.Getenv("key") == "" {
				log.Printf("The env variable with name %s not found, "+
					"nothing to delete\n", key)
			} else {
				err := os.Unsetenv(key)
				if err != nil {
					log.Printf("Can't unset env variable %s\n", key)
				}
			}
			continue
		}

		f, err := os.Open(filename)

		defer f.Close()

		if err != nil {
			log.Printf("Can't opent a file %s. %s\n", filename, err)
		}

		content, err := ioutil.ReadAll(f)
		c := string(content)
		if err != nil {
			log.Printf("Error reading file %s. %s\n", f, err)
		}

		log.Printf("[DEBUG] The content of file %s is %s\n", f, c)

		value := strings.TrimSpace(c)
		value = strings.TrimSuffix(value, "\t")
		value = strings.ReplaceAll(value, "0x00", "\n")
		valueSlice := strings.Split(value, "\n")
		value = valueSlice[0]

		log.Printf("[DEBUG] Try to set up env variable %s=%s", key, value)
		err = os.Setenv(key, value)

		if err != nil {
			log.Printf("Can't set env variable %s to value %s\n", key, value)
		}

		envValue := os.Getenv(key)
		log.Printf("The env variable %s was successfully set to %s", key, envValue)

	}
	cmdPath, err := exec.LookPath(command)
	if err != nil {
		log.Printf("The external command %s path not found. %s", command, err)
		return
	}
	log.Printf("Try to run external command %s on path %s", command, cmdPath)

	log.Printf("External command arguments is %s %s", arg1, arg2)

	cmd := exec.Command(cmdPath, arg1, arg2)
	out := bytes.NewBuffer([]byte{})
	cmd.Stdout = out
	err = cmd.Run()
	if err != nil {
		log.Printf("Failed to run external command %s", command)
	}

	if cmd.ProcessState.Success() {
		fmt.Println("External process run successfully with output: \n")
		fmt.Println(out.String())
	}

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		fmt.Println("External process finished with abnormal exit code %s",
			exitCode)
		os.Exit(exitCode)
	} else {
		fmt.Println("Program fully finished successfully")
		os.Exit(0)
	}

}
