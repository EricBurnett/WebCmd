package main

import (
	"github.com/EricBurnett/WebCmd/modules"
	"github.com/EricBurnett/WebCmd/resources"
	"html/template"
	"sort"
)

var MODULE_LIST_FILE = "templates/module_list.html.template"

type moduleListItem struct {
	Name     string
	Commands []string
}

type moduleList struct {
	Modules []moduleListItem
}

// Composes a nice list of all installed modules and their handlers, in HTML.
func ModuleList(m map[string]modules.Module) (template.HTML, error) {
	template_content, err := resources.Load(MODULE_LIST_FILE)
	if err != nil {
		return template.HTML(""), err
	}

	var moduleListTemplate = template.New("Module list template")
	moduleListTemplate, err = moduleListTemplate.Parse(string(template_content))
	if err != nil {
		return template.HTML(""), err
	}

	commandsByName := make(map[string][]string)
	for command, module := range m {
		if _, has := commandsByName[module.Name()]; !has {
			commandsByName[module.Name()] = make([]string, 0)
		}
		commandsByName[module.Name()] = append(commandsByName[module.Name()], command)
	}
	moduleList := &moduleList{make([]moduleListItem, 0)}
	names := make([]string, len(commandsByName))
	i := 0
	for k, _ := range commandsByName {
		names[i] = k
		i++
	}
	sort.Strings(names)
	for _, name := range names {
		sort.Strings(commandsByName[name])
		moduleList.Modules = append(moduleList.Modules,
			moduleListItem{name, commandsByName[name]})
	}
	var w modules.HTMLWriter
	moduleListTemplate.Execute(&w, &moduleList)
	return w.HTML(), nil
}
