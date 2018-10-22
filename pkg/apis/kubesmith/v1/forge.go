package v1

import (
	"encoding/json"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ForgeSpec defines the specification for a Kubesmith Forge.
type ForgeSpec struct {
	//
}

// ForgeStatus captures the current status of a Kubesmith Forge.
type ForgeStatus struct {
	LastUpdatedTime metav1.Time `json:"lastUpdatedTime"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Forge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   ForgeSpec   `json:"spec"`
	Status ForgeStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ForgeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Forge `json:"items"`
}

// helpers

func (p *Forge) GetPatchFromOriginal(original Forge) (types.PatchType, []byte, error) {
	p.Status.LastUpdatedTime.Time = time.Now()

	origBytes, err := json.Marshal(original)
	if err != nil {
		return "", nil, errors.Wrap(err, "error marshalling original forge")
	}

	updatedBytes, err := json.Marshal(p)
	if err != nil {
		return "", nil, errors.Wrap(err, "error marshalling updated forge")
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origBytes, updatedBytes)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating json merge patch for forge")
	}

	return types.MergePatchType, patchBytes, nil
}

func (p *Forge) Validate() error {
	// todo: finish this validation
	return nil
}
