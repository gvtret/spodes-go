package cosem

import (
	"encoding/asn1"
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

func TestAssociationLN_DefaultAttributes(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, err := NewAssociationLN(*obis)
	assert.NoError(t, err)

	attr3, err := associationLN.GetAttribute(3)
	assert.NoError(t, err)
	partners := attr3.(AssociatedPartnersID)
	assert.Equal(t, uint16(0), partners.ClientSAP)
	assert.Equal(t, uint16(0), partners.ServerSAP)

	attr4, err := associationLN.GetAttribute(4)
	assert.NoError(t, err)
	appCtx := attr4.(ApplicationContextName)
	assert.Equal(t, byte(0), appCtx.ContextID)
	assert.Equal(t, obis.String(), appCtx.LogicalName.String())

	attr8, err := associationLN.GetAttribute(8)
	assert.NoError(t, err)
	assert.Equal(t, AssociationStatusNonAssociated, attr8.(AssociationStatus))

	_, err = associationLN.GetAttribute(7)
	assert.ErrorIs(t, err, ErrAccessDenied)
}

func TestAssociationLN_WriteAttributes(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obis)

	mech := AuthenticationMechanismName{MechanismID: 1, MechanismName: asn1.ObjectIdentifier{2, 16, 756}}
	err := associationLN.SetAttribute(6, mech)
	assert.NoError(t, err)

	attr, err := associationLN.GetAttribute(6)
	assert.NoError(t, err)
	assert.Equal(t, mech, attr.(AuthenticationMechanismName))

	err = associationLN.SetAttribute(7, []byte{0x01, 0x02})
	assert.NoError(t, err)

	err = associationLN.SetAttribute(3, AssociatedPartnersID{ClientSAP: 1, ServerSAP: 16})
	assert.ErrorIs(t, err, ErrAccessDenied)
}

func TestAssociationLN_AssociationFlow(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obis)

	res, err := associationLN.Invoke(associationLNMethodAssociate, nil)
	assert.NoError(t, err)
	assert.Equal(t, AssociationStatusAssociationPending, res.(AssociationStatus))

	_, err = associationLN.Invoke(associationLNMethodAssociate, nil)
	assert.ErrorIs(t, err, ErrAccessDenied)

	_, err = associationLN.Invoke(associationLNMethodReplyToHLS, []interface{}{[]byte{0x01}})
	assert.NoError(t, err)

	attr8, err := associationLN.GetAttribute(8)
	assert.NoError(t, err)
	assert.Equal(t, AssociationStatusAssociated, attr8.(AssociationStatus))

	_, err = associationLN.Invoke(associationLNMethodReplyToHLS, []interface{}{[]byte{0x02}})
	assert.ErrorIs(t, err, ErrAccessDenied)
}

func TestAssociationLN_GetAssociationInformation(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obis)

	_, err := associationLN.Invoke(associationLNMethodGetAssociationInformation, nil)
	assert.ErrorIs(t, err, ErrAccessDenied)

	_, err = associationLN.Invoke(associationLNMethodAssociate, nil)
	assert.NoError(t, err)
	_, err = associationLN.Invoke(associationLNMethodReplyToHLS, []interface{}{[]byte{0x01}})
	assert.NoError(t, err)

	res, err := associationLN.Invoke(associationLNMethodGetAssociationInformation, nil)
	assert.NoError(t, err)
	info := res.(AssociationInformation)
	assert.Equal(t, AssociationStatusAssociated, associationLN.Attributes[8].Value.(AssociationStatus))
	assert.Equal(t, associationLN.Attributes[3].Value, info.AssociatedPartnersID)
	assert.Equal(t, associationLN.Attributes[4].Value, info.ApplicationContextName)
	assert.Equal(t, associationLN.Attributes[5].Value, info.XDLMSContextInfo)
	assert.Equal(t, associationLN.Attributes[6].Value, info.AuthenticationMechanism)
	assert.Equal(t, associationLN.Attributes[9].Value, info.SecuritySetupReference)
	assert.Equal(t, associationLN.Attributes[10].Value, info.ClientSAP)
	assert.Equal(t, associationLN.Attributes[11].Value, info.ServerSAP)
	assert.Equal(t, associationLN.Attributes[12].Value, info.UserList)
}

func TestAssociationLN_GetApplicationContextNameList(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obis)

	res, err := associationLN.Invoke(associationLNMethodGetApplicationContextNameList, nil)
	assert.NoError(t, err)
	list := res.([]ApplicationContextName)
	assert.Len(t, list, 1)
	assert.Equal(t, associationLN.Attributes[4].Value, list[0])
}

func TestAssociationLN_ReplyToHLSAuthentication_InvalidParameter(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obis)

	_, err := associationLN.Invoke(associationLNMethodAssociate, nil)
	assert.NoError(t, err)

	_, err = associationLN.Invoke(associationLNMethodReplyToHLS, []interface{}{"invalid"})
	assert.ErrorIs(t, err, ErrInvalidParameter)
}

func TestAssociationLN_ReplyToHLSAuthentication_Unauthorized(t *testing.T) {
	obis, _ := NewObisCodeFromString("0.0.40.0.0.255")
	associationLN, _ := NewAssociationLN(*obis)

	_, err := associationLN.Invoke(associationLNMethodReplyToHLS, []interface{}{[]byte{0x01}})
	assert.ErrorIs(t, err, ErrAccessDenied)
}
