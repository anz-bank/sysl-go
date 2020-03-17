package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/anz-bank/sysl-go/validator"

	"gopkg.in/yaml.v2"
)

func FindConfigFilename(cfgDir, prefix string) string {
	for _, ext := range []string{".json", ".yaml"} {
		filename := filepath.Join(cfgDir, prefix+ext)
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}
	return ""
}

func LoadAndValidateFromYAMLFileName(filename string, out validator.Validator) error {
	cfgBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = loadAndValidateFromYAML(cfgBytes, out)
	if err != nil {
		return fmt.Errorf("error encountered with YAML file %v: %v", filename, err)
	}

	return nil
}

func loadAndValidateFromYAML(b []byte, out validator.Validator) error {
	if err := yaml.UnmarshalStrict(b, out); err != nil {
		return fmt.Errorf("failed to parse: %v", err)
	}
	if err := out.Validate(); err != nil {
		return fmt.Errorf("failed to validate: %v", err)
	}

	return nil
}
