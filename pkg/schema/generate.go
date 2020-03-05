package schema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/client-go/kubernetes/scheme"
)

func Generate(customResourceDefinitionPath string, outputDir string) error {
	// we generate schemas from the config/crds in the root of this project
	// those crds can be created from controller-gen or by running `make openapischema`
	workdir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get workdir")
	}

	crdContents, err := readCRDFromPath(customResourceDefinitionPath)
	if err != nil {
		return errors.Wrap(err, "failed to read crd from path")
	}

	crdName, err := parseCRDName(crdContents)
	if err != nil {
		return errors.Wrap(err, "failed to parse crd name")
	}

	if err := generateSchemaFromCRD(crdContents, filepath.Join(workdir, outputDir, fmt.Sprintf("%s.json", crdName))); err != nil {
		return errors.Wrap(err, "failed to write schema")
	}

	return nil
}

func parseCRDName(crd []byte) (string, error) {
	extensionsscheme.AddToScheme(scheme.Scheme)
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(crd, nil, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse crd")
	}

	switch v := obj.(type) {
	case *extensionsv1.CustomResourceDefinition:
		return v.Name, nil
	case *extensionsv1beta1.CustomResourceDefinition:
		g := strings.Split(v.Spec.Group, ".")
		if len(g) == 0 {
			return "", errors.New("invalid group")
		}

		return fmt.Sprintf("%s-%s-%s", v.Spec.Names.Singular, g[0], v.Spec.Version), nil
	}

	return "", errors.New("data was not a crd")
}

func generateSchemaFromCRD(crd []byte, outfile string) error {
	extensionsscheme.AddToScheme(scheme.Scheme)
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(crd, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to decode crd")
	}

	customResourceDefinition := obj.(*extensionsv1beta1.CustomResourceDefinition)

	b, err := json.MarshalIndent(customResourceDefinition.Spec.Validation.OpenAPIV3Schema, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal json")
	}

	_, err = os.Stat(outfile)
	if err == nil {
		if err := os.Remove(outfile); err != nil {
			return errors.Wrap(err, "failed to remove file")
		}
	}

	d, _ := path.Split(outfile)
	_, err = os.Stat(d)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(d, 0755); err != nil {
			return errors.Wrap(err, "failed to mkdir")
		}
	}

	err = ioutil.WriteFile(outfile, b, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	return nil
}

func readCRDFromPath(specPath string) ([]byte, error) {
	if !isURL(specPath) {
		if _, err := os.Stat(specPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("%s was not found", specPath)
		}

		b, err := ioutil.ReadFile(specPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read file")
		}

		return b, nil
	}
	req, err := http.NewRequest("GET", specPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("User-Agent", "Replicated_CRDToOpenApiSchema/v1alpha1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	return body, nil
}

func isURL(str string) bool {
	parsed, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}

	return parsed.Scheme != ""
}
