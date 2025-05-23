// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/tview"
)

var (
	keyValRX = regexp.MustCompile(`\A(\s*)([\w\-./\s]+):\s(.+)\z`)
	keyRX    = regexp.MustCompile(`\A(\s*)([\w\-./\s]+):\s*\z`)
	searchRX = regexp.MustCompile(`<<<("search_\d+")>>>(.+)<<<"">>>`)
)

const (
	yamlFullFmt  = "%s[key::b]%s[colon::-]: [val::]%s"
	yamlKeyFmt   = "%s[key::b]%s[colon::-]:"
	yamlValueFmt = "[val::]%s"
)

func colorizeYAML(style config.Yaml, raw string) string {
	lines := strings.Split(tview.Escape(raw), "\n")
	fullFmt := strings.Replace(yamlFullFmt, "[key", "["+style.KeyColor.String(), 1)
	fullFmt = strings.Replace(fullFmt, "[colon", "["+style.ColonColor.String(), 1)
	fullFmt = strings.Replace(fullFmt, "[val", "["+style.ValueColor.String(), 1)

	keyFmt := strings.Replace(yamlKeyFmt, "[key", "["+style.KeyColor.String(), 1)
	keyFmt = strings.Replace(keyFmt, "[colon", "["+style.ColonColor.String(), 1)

	valFmt := strings.Replace(yamlValueFmt, "[val", "["+style.ValueColor.String(), 1)

	buff := make([]string, 0, len(lines))
	for _, l := range lines {
		res := keyValRX.FindStringSubmatch(l)
		if len(res) == 4 {
			buff = append(buff, enableRegion(fmt.Sprintf(fullFmt, res[1], res[2], res[3])))
			continue
		}

		res = keyRX.FindStringSubmatch(l)
		if len(res) == 3 {
			buff = append(buff, enableRegion(fmt.Sprintf(keyFmt, res[1], res[2])))
			continue
		}

		buff = append(buff, enableRegion(fmt.Sprintf(valFmt, l)))
	}

	return strings.Join(buff, "\n")
}

func enableRegion(s string) string {
	if searchRX.MatchString(s) {
		return strings.ReplaceAll(strings.ReplaceAll(s, "<<<", "["), ">>>", "]")
	}

	return s
}

func saveYAML(dir, name, raw string) (string, error) {
	if err := ensureDir(dir); err != nil {
		return "", err
	}

	fName := fmt.Sprintf("%s--%d.yaml", data.SanitizeFileName(name), time.Now().Unix())
	fpath := filepath.Join(dir, fName)
	mod := os.O_CREATE | os.O_WRONLY
	file, err := os.OpenFile(fpath, mod, 0600)
	if err != nil {
		slog.Error("Unable to open YAML file",
			slogs.Path, fpath,
			slogs.Error, err,
		)
		return "", nil
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("Closing yaml file failed",
				slogs.Path, fpath,
				slogs.Error, err,
			)
		}
	}()
	if _, err := file.WriteString(raw); err != nil {
		return "", err
	}

	return fpath, nil
}
