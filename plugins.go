package main

import (
	"os"
	"io/ioutil"
	"github.com/smallfish/simpleyaml"
)

func LoadPlugins() {
	files, _ := ioutil.ReadDir("./plugins")
    for _, f := range files {
		if f.IsDir() {
    		LoadPlugin(f.Name())
		}
    }
}

func LoadPlugin(name string) {
	configFile := "./plugins/" + name + "/plugin.yaml"
	
	if _, err := os.Stat(configFile); err != nil {
		warnf("Plugin missing configuration: %s", name)
		return
	}
	
	configData, err := ioutil.ReadFile(configFile)
	
	if err != nil {
		warnf("error opening config: %v", err)
	}
    
    config, err := simpleyaml.NewYaml(configData)
    
    title, err := config.Get("plugin").String()
    author, err := config.Get("author").String()
    description, err := config.Get("description").String()
	
//    infof("plugin: %v", config);
    infof("plugin name: %s", title);
    infof("plugin author: %s", author);
    infof("plugin description: %s", description);
    
    commands, err := config.Get("commands").Map()
    info("Commands:")
    
    for key, _ := range commands {
        key, _ := key.(string)
        command := config.GetPath("commands", key)
        textCommand, _ := command.Get("command").String()
        infof("!%s", textCommand)
    }
}
