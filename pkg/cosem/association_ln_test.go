package cosem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssociationLN_AddObject(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obis)

	obisData, _ := NewObisCodeFromString("1.0.0.3.0.255")
	dataObj, _ := NewData(*obisData, uint32(12345))

	associationLN.AddObject(dataObj)

	objList := associationLN.Attributes[2].Value.([]ObjectListElement)
	assert.Len(t, objList, 1)
	assert.Equal(t, DataClassID, objList[0].ClassID)
	assert.Equal(t, *obisData, objList[0].InstanceID)
}
