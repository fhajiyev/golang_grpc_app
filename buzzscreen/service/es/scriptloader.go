package es

import (
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/Buzzvil/buzzlib-go/core"
	"gopkg.in/olivere/elastic.v5"
)

// JoinOp type definition
type JoinOp string

// JoinOp constants
const (
	JoinOpAdd      JoinOp = "+"
	JoinOpMultiply JoinOp = "*"
)

// DefaultModelArtifact to load if requested script file doesn't exist
const DefaultModelArtifact string = "v4_a"

// ScriptLoader type definition
type ScriptLoader struct {
	toPath  string // should be ".../service"
	scripts map[string]string

	filterScript *string
}

var l sync.Mutex // Mutex Lock for sync file access

// GetScriptFields returns es.ScriptFields to be appended to searchSource for debugging
func (sl *ScriptLoader) GetScriptFields(modelArtifact string, params map[string]interface{}) []*elastic.ScriptField {
	modelArtifact = sl.GetRunnableModelArtifact(modelArtifact)
	filepath := path.Join(sl.toPath, fmt.Sprintf("es/script/%s/", modelArtifact))
	files, _ := ioutil.ReadDir(filepath)

	var sf []*elastic.ScriptField
	for _, f := range files {
		scriptName := strings.Split(f.Name(), ".")[0]
		script := elastic.NewScript(sl.GetFieldScript(modelArtifact, scriptName)).Params(params)
		sf = append(sf, elastic.NewScriptField(scriptName, script))
	}
	return sf
}

// GetFilterScript returns a preloaded script or loads new script with the name
func (sl *ScriptLoader) GetFilterScript() string {
	if sl.filterScript == nil {
		filepath := path.Join(sl.toPath, "es/script/filter/filter.painless")
		b, err := ioutil.ReadFile(filepath)
		if err != nil {
			core.Logger.Errorf("GetFilterScript() - filter script file doesn't exist: %v", err)
		}

		filterScript := string(b)
		sl.filterScript = &filterScript
	}

	return *sl.filterScript
}

// GetFieldScript returns a preloaded script or loads new script with the name
func (sl *ScriptLoader) GetFieldScript(modelArtifact string, scriptName string) string {
	scriptKey := fmt.Sprintf("%s-%s", modelArtifact, scriptName)
	if script, ok := sl.scripts[scriptKey]; ok {
		return script
	}
	l.Lock()         // lock
	defer l.Unlock() // unlock

	filepath := path.Join(sl.toPath, fmt.Sprintf("es/script/%s/%s.painless", modelArtifact, scriptName))
	b, err := ioutil.ReadFile(filepath)

	if err != nil {
		core.Logger.Errorf("GetFieldScript(%v) - File doesn't exist: %v", scriptName, err)
	}
	sl.scripts[scriptKey] = string(b)

	return sl.scripts[scriptKey]
}

// GetScoreScript returns combined score script using intermediate scripts
func (sl *ScriptLoader) GetScoreScript(modelArtifact string, jop JoinOp) string {
	if script, ok := sl.scripts[modelArtifact]; ok {
		return script
	}

	modelArtifact = sl.GetRunnableModelArtifact(modelArtifact)
	filepath := path.Join(sl.toPath, fmt.Sprintf("es/script/%s/", modelArtifact))
	files, _ := ioutil.ReadDir(filepath)

	var scriptLines []string
	var factors []string
	for _, f := range files {
		scriptName := strings.Split(f.Name(), ".")[0]
		lines := strings.Split(sl.GetFieldScript(modelArtifact, scriptName), "\n")
		scriptLines = append(scriptLines, lines[0:len(lines)-1]...)
		factors = append(factors, lines[len(lines)-1])
	}

	scriptLines = append(scriptLines, strings.Join(factors, string(jop)))
	sl.scripts[modelArtifact] = strings.Join(scriptLines, "\n")
	return sl.scripts[modelArtifact]
}

// GetRunnableModelArtifact func definition
func (sl *ScriptLoader) GetRunnableModelArtifact(requestedModelArtifact string) string {
	filepath := path.Join(sl.toPath, fmt.Sprintf("es/script/%s/", requestedModelArtifact))
	_, err := ioutil.ReadDir(filepath)
	if err != nil {
		core.Logger.Warnf("GetRunnableModelArtifact() - Failed reading modelArtifact dir %v - %v - Using %v instead!", requestedModelArtifact, err, DefaultModelArtifact)
		return DefaultModelArtifact
	}
	return requestedModelArtifact
}

var scriptLoaderInstance *ScriptLoader

// GetScriptLoader returns ScriptLoader singleton instance
func GetScriptLoader() *ScriptLoader {
	if scriptLoaderInstance == nil {
		_, filename, _, _ := runtime.Caller(1)
		scriptLoaderInstance = &ScriptLoader{scripts: make(map[string]string), toPath: path.Dir(filename)}
	}
	return scriptLoaderInstance
}
