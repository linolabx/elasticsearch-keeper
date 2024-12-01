package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samber/lo"
)

func escapeSynonyms(synonyms string) string {
	return strings.ReplaceAll(strings.ReplaceAll(synonyms, " ", "\\\\ "), ",", "\\\\,")
}

type Response struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func (ek *EsKeeper) SetSynonyms(synonymsFileName string, synonymList *[]*[]string, indexes ...string) error {
	if !strings.HasPrefix(synonymsFileName, "synonyms/") {
		return fmt.Errorf("synonyms files must be placed in the synonyms/ directory")
	}

	synonymsRaw := strings.Join(lo.Map(*synonymList, func(synonym *[]string, _ int) string {
		return strings.Join(lo.Map(*synonym, func(word string, _ int) string {
			return escapeSynonyms(word)
		}), ",")
	}), "\n")

	request := ek.resty.R()
	request.SetMultipartField("file", synonymsFileName, "text/plain", bytes.NewReader([]byte(synonymsRaw)))

	if len(indexes) > 0 {
		for _, index := range indexes {
			request.FormData.Add("indexes[]", index)
		}
	}

	resp, err := request.Post(synonymsFileName)
	if err != nil {
		return err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return fmt.Errorf("failed to parse response: %s", err)
	}

	if response.Error != "" {
		return fmt.Errorf("failed to set synonyms: %s", response.Error)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to set synonyms: %s", resp.String())
	}

	return nil
}
