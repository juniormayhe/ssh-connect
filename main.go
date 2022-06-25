package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"os/exec"
	"os/signal"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Credential struct {
	Name          string
	Path          string
	UserAndServer string
}

type Config struct {
	//Credentials map[string][]Credential
	Credentials []Credential
}

func main() {
	fmt.Println("ssh-connect\n")

	config, err := readConfig()
	if err != nil {
		fmt.Printf("Failed to read config. %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("Choose an option:")

	for k, v := range config.Credentials {
		fmt.Printf("- %v %v\n", k+1, v.Name)
	}

	var n int
	fmt.Scan(&n)
	if err != nil {
		fmt.Printf("Failed to read input. %s\n", err.Error())
		os.Exit(1)
	}

	n -= 1
	fmt.Println()

	path := fmt.Sprintf("%v", config.Credentials[n].Path)
	path = normalizePath(path)
	userAndServer := fmt.Sprintf("%v", config.Credentials[n].UserAndServer)

	var cmd *exec.Cmd
	if getPathFormat() == "linux" {
		cmd = exec.Command("bash", "-c", "ssh -i "+path+" "+userAndServer+" -o StrictHostKeyChecking=no")
	} else {
		cmd = exec.Command("ssh", "-i", path, userAndServer, "-o", "StrictHostKeyChecking=no")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Printf("captured %v, stopping...", sig)
			stdout, _ := cmd.CombinedOutput()
			fmt.Printf("Stopped. %s\n", string(stdout))
			os.Exit(1)

		}
	}()

	err = cmd.Run()

	if err != nil {
		fmt.Printf("Failed to start bash. %s\n", err.Error())
		os.Exit(1)
	} else {
		fmt.Println("Done!")
	}

}

func readConfig() (Config, error) {
	filename, _ := filepath.Abs("./ssh-keys.yaml")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Printf("Failed to read config. %s\n", err.Error())
		return Config{}, err
	}

	var c Config

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		fmt.Printf("Failed to parse config. %s\n", err.Error())
		return Config{}, err
	}
	return c, err
}

func getPathFormat() string {
	currentPath, _ := exec.Command("pwd").Output()
	if strings.Contains(string(currentPath), "/") {
		return "linux"
	}
	return "windows"
}

func normalizePath(path string) string {
	newpath := ""
	format := getPathFormat()
	if format == "linux" {
		newpath = strings.Replace(path, "\\", "/", -1)
		//TODO: regex to get any disk unit letter
		newpath = strings.Replace(newpath, "C:", "/C", -1)
		fmt.Printf("SSH Keys paths are now in Linux format %s\n", newpath)
	} else {
		newpath = strings.Replace(path, "/C", "C:", -1)
		//TODO: regex to get any disk unit letter
		newpath = strings.Replace(newpath, "/", "\\", -1)
		fmt.Printf("SSH Keys paths are now in Windows format %s\n", newpath)
	}

	return newpath

}
