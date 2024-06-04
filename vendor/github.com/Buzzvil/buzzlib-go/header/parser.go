package header

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Parser defines struct for header parser
type Parser struct{}

// New returns headerParser
func New() *Parser {
	return &Parser{}
}

const (
	sdkRegex                = `([\w.-]+)?\/([\w.-]+)\s*\((\d*)?\)`
	appRegex                = `([\w.-]+)?\/([\w.-]+)\s*\((\d*)?\)`
	osRegex                 = `(\w+)?\/([\w.-]+)`
	extraDeviceRegex        = `\((\d*)?;(.*)?;(.*)?;(.*)?;(.*)?;(.*)?\)`
	buzzUserAgentParseRegex = `^` + sdkRegex + `\s*` + appRegex + `\s*` + osRegex + `\s*` + extraDeviceRegex + `$`
)

// ParseUserAgent parses user agent from a string
func (p *Parser) ParseUserAgent(userAgentStr string) (*BuzzUserAgent, error) {
	re, err := regexp.Compile(buzzUserAgentParseRegex)
	if err != nil {
		return nil, fmt.Errorf("ParseUserAgent regex error %v", err)
	}

	// If it does not matches reges then, return string array is empty
	// Else, it always return string array which has length of 15
	match := re.FindStringSubmatch(userAgentStr)
	if len(match) == 0 {
		return nil, fmt.Errorf("ParseUserAgent not found %s", userAgentStr)
	}

	return &BuzzUserAgent{
		SDKName:        match[1],
		SDKVersion:     match[2],
		SDKVersionCode: p.getVersionCode(match[3]),
		AppName:        match[4],
		AppVersion:     match[5],
		AppVersionCode: p.getVersionCode(match[6]),
		OS:             match[7],
		OSVersion:      match[8],
		OSVersionCode:  p.getVersionCode(match[9]),
		Model:          match[10],
		Manufacturer:   match[11],
		Device:         match[12],
		Brand:          match[13],
		Product:        match[14],
	}, nil
}

func (p *Parser) getVersionCode(versionWrapped string) int64 {
	verCodeStr := strings.TrimFunc(versionWrapped, func(r rune) bool {
		return r == '(' || r == ')'
	})

	if verCodeStr == "" {
		return 0
	}

	verCode, err := strconv.ParseInt(verCodeStr, 10, 64)
	if err != nil {
		return 0
	}

	return verCode
}
