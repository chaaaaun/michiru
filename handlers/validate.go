//go:build cgo

package handlers

import (
	_ "embed"
	"fmt"
	"time"

	xsdvalidate "github.com/terminalstatic/go-xsd-validate"
	"michiru/models"
)

//go:embed schema.xsd
var schema []byte

func ValidateXml(b []byte) error {
	err := xsdvalidate.Init()
	if err != nil {
		return err
	}
	defer xsdvalidate.Cleanup()

	xsdhandler, err := xsdvalidate.NewXsdHandlerMem(schema, xsdvalidate.ParsErrVerbose)
	if err != nil {
		return err
	}
	defer xsdhandler.Free()

	err = xsdhandler.ValidateMem(b, xsdvalidate.ValidErrDefault)
	if err != nil {
		return err
	}

	return nil
}

func ValidateImportInterval(meta *models.MetadataDocument) error {
	// We assume no metadata means its the first ever import, so we can skip this check
	if meta == nil {
		return nil
	}

	dayBefore := time.Now().Add(-24 * time.Hour)
	if meta.RetrievedAt.After(dayBefore) {
		return fmt.Errorf("last import at %s, less than a day ago", meta.RetrievedAt)
	}

	return nil
}
