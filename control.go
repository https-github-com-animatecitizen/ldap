package ldap

import (
	"fmt"
	"strconv"

	ber "github.com/go-asn1-ber/asn1-ber"
)

const (
	// ControlTypePaging - https://www.ietf.org/rfc/rfc2696.txt
	ControlTypePaging = "1.2.840.113556.1.4.319"
	// ControlTypeBeheraPasswordPolicy - https://tools.ietf.org/html/draft-behera-ldap-password-policy-10
	ControlTypeBeheraPasswordPolicy = "1.3.6.1.4.1.42.2.27.8.5.1"
	// ControlTypeVChuPasswordMustChange - https://tools.ietf.org/html/draft-vchu-ldap-pwd-policy-00
	ControlTypeVChuPasswordMustChange = "2.16.840.1.113730.3.4.4"
	// ControlTypeVChuPasswordWarning - https://tools.ietf.org/html/draft-vchu-ldap-pwd-policy-00
	ControlTypeVChuPasswordWarning = "2.16.840.1.113730.3.4.5"
	// ControlTypeManageDsaIT - https://tools.ietf.org/html/rfc3296
	ControlTypeManageDsaIT = "2.16.840.1.113730.3.4.2"
	// ControlTypeWhoAmI - https://tools.ietf.org/html/rfc4532
	ControlTypeWhoAmI = "1.3.6.1.4.1.4203.1.11.3"
	// ControlTypeSubtreeDelete - https://datatracker.ietf.org/doc/html/draft-armijo-ldap-treedelete-02
	ControlTypeSubtreeDelete = "1.2.840.113556.1.4.805"

	// ControlTypeMicrosoftNotification - https://msdn.microsoft.com/en-us/library/aa366983(v=vs.85).aspx
	ControlTypeMicrosoftNotification = "1.2.840.113556.1.4.528"
	// ControlTypeMicrosoftShowDeleted - https://msdn.microsoft.com/en-us/library/aa366989(v=vs.85).aspx
	ControlTypeMicrosoftShowDeleted = "1.2.840.113556.1.4.417"
	// ControlTypeMicrosoftServerLinkTTL - https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/f4f523a8-abc0-4b3a-a471-6b2fef135481?redirectedfrom=MSDN
	ControlTypeMicrosoftServerLinkTTL = "1.2.840.113556.1.4.2309"
)

// ControlTypeMap maps controls to text descriptions
var ControlTypeMap = map[string]string{
	ControlTypePaging:                 "Paging",
	ControlTypeBeheraPasswordPolicy:   "Password Policy - Behera Draft",
	ControlTypeManageDsaIT:            "Manage DSA IT",
	ControlTypeSubtreeDelete:          "Subtree Delete Control",
	ControlTypeMicrosoftNotification:  "Change Notification - Microsoft",
	ControlTypeMicrosoftShowDeleted:   "Show Deleted Objects - Microsoft",
	ControlTypeMicrosoftServerLinkTTL: "Return TTL-DNs for link values with associated expiry times - Microsoft",
}

// Control defines an interface controls provide to encode and describe themselves
type Control interface {
	// GetControlType returns the OID
	GetControlType() string
	// Encode returns the ber packet representation
	Encode() *ber.Packet
	// String returns a human-readable description
	String() string
}

// ControlString implements the Control interface for simple controls
type ControlString struct {
	ControlType  string
	Criticality  bool
	ControlValue string
}

// GetControlType returns the OID
func (c *ControlString) GetControlType() string {
	return c.ControlType
}

// Encode returns the ber packet representation
func (c *ControlString) Encode() *ber.Packet {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, c.ControlType, "Control Type ("+ControlTypeMap[c.ControlType]+")"))
	if c.Criticality {
		packet.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, c.Criticality, "Criticality"))
	}
	if c.ControlValue != "" {
		packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, string(c.ControlValue), "Control Value"))
	}
	return packet
}

// String returns a human-readable description
func (c *ControlString) String() string {
	return fmt.Sprintf("Control Type: %s (%q)  Criticality: %t  Control Value: %s", ControlTypeMap[c.ControlType], c.ControlType, c.Criticality, c.ControlValue)
}

// ControlPaging implements the paging control described in https://www.ietf.org/rfc/rfc2696.txt
type ControlPaging struct {
	// PagingSize indicates the page size
	PagingSize uint32
	// Cookie is an opaque value returned by the server to track a paging cursor
	Cookie []byte
}

// GetControlType returns the OID
func (c *ControlPaging) GetControlType() string {
	return ControlTypePaging
}

// Encode returns the ber packet representation
func (c *ControlPaging) Encode() *ber.Packet {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, ControlTypePaging, "Control Type ("+ControlTypeMap[ControlTypePaging]+")"))

	p2 := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Control Value (Paging)")
	seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Search Control Value")
	seq.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, int64(c.PagingSize), "Paging Size"))
	cookie := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Cookie")
	cookie.Value = c.Cookie
	cookie.Data.Write(c.Cookie)
	seq.AppendChild(cookie)
	p2.AppendChild(seq)

	packet.AppendChild(p2)
	return packet
}

// String returns a human-readable description
func (c *ControlPaging) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)  Criticality: %t  PagingSize: %d  Cookie: %q",
		ControlTypeMap[ControlTypePaging],
		ControlTypePaging,
		false,
		c.PagingSize,
		c.Cookie)
}

// SetCookie stores the given cookie in the paging control
func (c *ControlPaging) SetCookie(cookie []byte) {
	c.Cookie = cookie
}

// ControlBeheraPasswordPolicy implements the control described in https://tools.ietf.org/html/draft-behera-ldap-password-policy-10
type ControlBeheraPasswordPolicy struct {
	// Expire contains the number of seconds before a password will expire
	Expire int64
	// Grace indicates the remaining number of times a user will be allowed to authenticate with an expired password
	Grace int64
	// Error indicates the error code
	Error int8
	// ErrorString is a human readable error
	ErrorString string
}

// GetControlType returns the OID
func (c *ControlBeheraPasswordPolicy) GetControlType() string {
	return ControlTypeBeheraPasswordPolicy
}

// Encode returns the ber packet representation
func (c *ControlBeheraPasswordPolicy) Encode() *ber.Packet {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, ControlTypeBeheraPasswordPolicy, "Control Type ("+ControlTypeMap[ControlTypeBeheraPasswordPolicy]+")"))

	return packet
}

// String returns a human-readable description
func (c *ControlBeheraPasswordPolicy) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)  Criticality: %t  Expire: %d  Grace: %d  Error: %d, ErrorString: %s",
		ControlTypeMap[ControlTypeBeheraPasswordPolicy],
		ControlTypeBeheraPasswordPolicy,
		false,
		c.Expire,
		c.Grace,
		c.Error,
		c.ErrorString)
}

// ControlVChuPasswordMustChange implements the control described in https://tools.ietf.org/html/draft-vchu-ldap-pwd-policy-00
type ControlVChuPasswordMustChange struct {
	// MustChange indicates if the password is required to be changed
	MustChange bool
}

// GetControlType returns the OID
func (c *ControlVChuPasswordMustChange) GetControlType() string {
	return ControlTypeVChuPasswordMustChange
}

// Encode returns the ber packet representation
func (c *ControlVChuPasswordMustChange) Encode() *ber.Packet {
	return nil
}

// String returns a human-readable description
func (c *ControlVChuPasswordMustChange) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)  Criticality: %t  MustChange: %v",
		ControlTypeMap[ControlTypeVChuPasswordMustChange],
		ControlTypeVChuPasswordMustChange,
		false,
		c.MustChange)
}

// ControlVChuPasswordWarning implements the control described in https://tools.ietf.org/html/draft-vchu-ldap-pwd-policy-00
type ControlVChuPasswordWarning struct {
	// Expire indicates the time in seconds until the password expires
	Expire int64
}

// GetControlType returns the OID
func (c *ControlVChuPasswordWarning) GetControlType() string {
	return ControlTypeVChuPasswordWarning
}

// Encode returns the ber packet representation
func (c *ControlVChuPasswordWarning) Encode() *ber.Packet {
	return nil
}

// String returns a human-readable description
func (c *ControlVChuPasswordWarning) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)  Criticality: %t  Expire: %b",
		ControlTypeMap[ControlTypeVChuPasswordWarning],
		ControlTypeVChuPasswordWarning,
		false,
		c.Expire)
}

// ControlManageDsaIT implements the control described in https://tools.ietf.org/html/rfc3296
type ControlManageDsaIT struct {
	// Criticality indicates if this control is required
	Criticality bool
}

// GetControlType returns the OID
func (c *ControlManageDsaIT) GetControlType() string {
	return ControlTypeManageDsaIT
}

// Encode returns the ber packet representation
func (c *ControlManageDsaIT) Encode() *ber.Packet {
	// FIXME
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, ControlTypeManageDsaIT, "Control Type ("+ControlTypeMap[ControlTypeManageDsaIT]+")"))
	if c.Criticality {
		packet.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, c.Criticality, "Criticality"))
	}
	return packet
}

// String returns a human-readable description
func (c *ControlManageDsaIT) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)  Criticality: %t",
		ControlTypeMap[ControlTypeManageDsaIT],
		ControlTypeManageDsaIT,
		c.Criticality)
}

// NewControlManageDsaIT returns a ControlManageDsaIT control
func NewControlManageDsaIT(Criticality bool) *ControlManageDsaIT {
	return &ControlManageDsaIT{Criticality: Criticality}
}

// ControlMicrosoftNotification implements the control described in https://msdn.microsoft.com/en-us/library/aa366983(v=vs.85).aspx
type ControlMicrosoftNotification struct{}

// GetControlType returns the OID
func (c *ControlMicrosoftNotification) GetControlType() string {
	return ControlTypeMicrosoftNotification
}

// Encode returns the ber packet representation
func (c *ControlMicrosoftNotification) Encode() *ber.Packet {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, ControlTypeMicrosoftNotification, "Control Type ("+ControlTypeMap[ControlTypeMicrosoftNotification]+")"))

	return packet
}

// String returns a human-readable description
func (c *ControlMicrosoftNotification) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)",
		ControlTypeMap[ControlTypeMicrosoftNotification],
		ControlTypeMicrosoftNotification)
}

// NewControlMicrosoftNotification returns a ControlMicrosoftNotification control
func NewControlMicrosoftNotification() *ControlMicrosoftNotification {
	return &ControlMicrosoftNotification{}
}

// ControlMicrosoftShowDeleted implements the control described in https://msdn.microsoft.com/en-us/library/aa366989(v=vs.85).aspx
type ControlMicrosoftShowDeleted struct{}

// GetControlType returns the OID
func (c *ControlMicrosoftShowDeleted) GetControlType() string {
	return ControlTypeMicrosoftShowDeleted
}

// Encode returns the ber packet representation
func (c *ControlMicrosoftShowDeleted) Encode() *ber.Packet {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, ControlTypeMicrosoftShowDeleted, "Control Type ("+ControlTypeMap[ControlTypeMicrosoftShowDeleted]+")"))

	return packet
}

// String returns a human-readable description
func (c *ControlMicrosoftShowDeleted) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)",
		ControlTypeMap[ControlTypeMicrosoftShowDeleted],
		ControlTypeMicrosoftShowDeleted)
}

// NewControlMicrosoftShowDeleted returns a ControlMicrosoftShowDeleted control
func NewControlMicrosoftShowDeleted() *ControlMicrosoftShowDeleted {
	return &ControlMicrosoftShowDeleted{}
}

// ControlMicrosoftServerLinkTTL implements the control described in https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/f4f523a8-abc0-4b3a-a471-6b2fef135481?redirectedfrom=MSDN
type ControlMicrosoftServerLinkTTL struct{}

// GetControlType returns the OID
func (c *ControlMicrosoftServerLinkTTL) GetControlType() string {
	return ControlTypeMicrosoftServerLinkTTL
}

// Encode returns the ber packet representation
func (c *ControlMicrosoftServerLinkTTL) Encode() *ber.Packet {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, ControlTypeMicrosoftServerLinkTTL, "Control Type ("+ControlTypeMap[ControlTypeMicrosoftServerLinkTTL]+")"))

	return packet
}

// String returns a human-readable description
func (c *ControlMicrosoftServerLinkTTL) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)",
		ControlTypeMap[ControlTypeMicrosoftServerLinkTTL],
		ControlTypeMicrosoftServerLinkTTL)
}

// NewControlMicrosoftServerLinkTTL returns a ControlMicrosoftServerLinkTTL control
func NewControlMicrosoftServerLinkTTL() *ControlMicrosoftServerLinkTTL {
	return &ControlMicrosoftServerLinkTTL{}
}

// FindControl returns the first control of the given type in the list, or nil
func FindControl(controls []Control, controlType string) Control {
	for _, c := range controls {
		if c.GetControlType() == controlType {
			return c
		}
	}
	return nil
}

// DecodeControl returns a control read from the given packet, or nil if no recognized control can be made
func DecodeControl(packet *ber.Packet) (Control, error) {
	var (
		ControlType = ""
		Criticality = false
		value       *ber.Packet
	)

	switch len(packet.Children) {
	case 0:
		// at least one child is required for control type
		return nil, fmt.Errorf("at least one child is required for control type")

	case 1:
		// just type, no criticality or value
		packet.Children[0].Description = "Control Type (" + ControlTypeMap[ControlType] + ")"
		ControlType = packet.Children[0].Value.(string)

	case 2:
		packet.Children[0].Description = "Control Type (" + ControlTypeMap[ControlType] + ")"
		ControlType = packet.Children[0].Value.(string)

		// Children[1] could be criticality or value (both are optional)
		// duck-type on whether this is a boolean
		if _, ok := packet.Children[1].Value.(bool); ok {
			packet.Children[1].Description = "Criticality"
			Criticality = packet.Children[1].Value.(bool)
		} else {
			packet.Children[1].Description = "Control Value"
			value = packet.Children[1]
		}

	case 3:
		packet.Children[0].Description = "Control Type (" + ControlTypeMap[ControlType] + ")"
		ControlType = packet.Children[0].Value.(string)

		packet.Children[1].Description = "Criticality"
		Criticality = packet.Children[1].Value.(bool)

		packet.Children[2].Description = "Control Value"
		value = packet.Children[2]

	default:
		// more than 3 children is invalid
		return nil, fmt.Errorf("more than 3 children is invalid for controls")
	}

	switch ControlType {
	case ControlTypeManageDsaIT:
		return NewControlManageDsaIT(Criticality), nil
	case ControlTypePaging:
		value.Description += " (Paging)"
		c := new(ControlPaging)
		if value.Value != nil {
			valueChildren, err := ber.DecodePacketErr(value.Data.Bytes())
			if err != nil {
				return nil, fmt.Errorf("failed to decode data bytes: %s", err)
			}
			value.Data.Truncate(0)
			value.Value = nil
			value.AppendChild(valueChildren)
		}
		value = value.Children[0]
		value.Description = "Search Control Value"
		value.Children[0].Description = "Paging Size"
		value.Children[1].Description = "Cookie"
		c.PagingSize = uint32(value.Children[0].Value.(int64))
		c.Cookie = value.Children[1].Data.Bytes()
		value.Children[1].Value = c.Cookie
		return c, nil
	case ControlTypeBeheraPasswordPolicy:
		value.Description += " (Password Policy - Behera)"
		c := NewControlBeheraPasswordPolicy()
		if value.Value != nil {
			valueChildren, err := ber.DecodePacketErr(value.Data.Bytes())
			if err != nil {
				return nil, fmt.Errorf("failed to decode data bytes: %s", err)
			}
			value.Data.Truncate(0)
			value.Value = nil
			value.AppendChild(valueChildren)
		}

		sequence := value.Children[0]

		for _, child := range sequence.Children {
			if child.Tag == 0 {
				// Warning
				warningPacket := child.Children[0]
				val, err := ber.ParseInt64(warningPacket.Data.Bytes())
				if err != nil {
					return nil, fmt.Errorf("failed to decode data bytes: %s", err)
				}
				if warningPacket.Tag == 0 {
					// timeBeforeExpiration
					c.Expire = val
					warningPacket.Value = c.Expire
				} else if warningPacket.Tag == 1 {
					// graceAuthNsRemaining
					c.Grace = val
					warningPacket.Value = c.Grace
				}
			} else if child.Tag == 1 {
				// Error
				bs := child.Data.Bytes()
				if len(bs) != 1 || bs[0] > 8 {
					return nil, fmt.Errorf("failed to decode data bytes: %s", "invalid PasswordPolicyResponse enum value")
				}
				val := int8(bs[0])
				c.Error = val
				child.Value = c.Error
				c.ErrorString = BeheraPasswordPolicyErrorMap[c.Error]
			}
		}
		return c, nil
	case ControlTypeVChuPasswordMustChange:
		c := &ControlVChuPasswordMustChange{MustChange: true}
		return c, nil
	case ControlTypeVChuPasswordWarning:
		c := &ControlVChuPasswordWarning{Expire: -1}
		expireStr := ber.DecodeString(value.Data.Bytes())

		expire, err := strconv.ParseInt(expireStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value as int: %s", err)
		}
		c.Expire = expire
		value.Value = c.Expire

		return c, nil
	case ControlTypeMicrosoftNotification:
		return NewControlMicrosoftNotification(), nil
	case ControlTypeMicrosoftShowDeleted:
		return NewControlMicrosoftShowDeleted(), nil
	case ControlTypeMicrosoftServerLinkTTL:
		return NewControlMicrosoftServerLinkTTL(), nil
	case ControlTypeSubtreeDelete:
		return NewControlSubtreeDelete(), nil
	default:
		c := new(ControlString)
		c.ControlType = ControlType
		c.Criticality = Criticality
		if value != nil {
			c.ControlValue = value.Value.(string)
		}
		return c, nil
	}
}

// NewControlString returns a generic control
func NewControlString(controlType string, criticality bool, controlValue string) *ControlString {
	return &ControlString{
		ControlType:  controlType,
		Criticality:  criticality,
		ControlValue: controlValue,
	}
}

// NewControlPaging returns a paging control
func NewControlPaging(pagingSize uint32) *ControlPaging {
	return &ControlPaging{PagingSize: pagingSize}
}

// NewControlBeheraPasswordPolicy returns a ControlBeheraPasswordPolicy
func NewControlBeheraPasswordPolicy() *ControlBeheraPasswordPolicy {
	return &ControlBeheraPasswordPolicy{
		Expire: -1,
		Grace:  -1,
		Error:  -1,
	}
}

// ControlSubtreeDelete implements the subtree delete control described in
// https://datatracker.ietf.org/doc/html/draft-armijo-ldap-treedelete-02
type ControlSubtreeDelete struct{}

// GetControlType returns the OID
func (c *ControlSubtreeDelete) GetControlType() string {
	return ControlTypeSubtreeDelete
}

// NewControlSubtreeDelete returns a ControlSubtreeDelete control.
func NewControlSubtreeDelete() *ControlSubtreeDelete {
	return &ControlSubtreeDelete{}
}

// Encode returns the ber packet representation
func (c *ControlSubtreeDelete) Encode() *ber.Packet {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, ControlTypeSubtreeDelete, "Control Type ("+ControlTypeMap[ControlTypeSubtreeDelete]+")"))

	return packet
}

func (c *ControlSubtreeDelete) String() string {
	return fmt.Sprintf(
		"Control Type: %s (%q)",
		ControlTypeMap[ControlTypeSubtreeDelete],
		ControlTypeSubtreeDelete)
}

func encodeControls(controls []Control) *ber.Packet {
	packet := ber.Encode(ber.ClassContext, ber.TypeConstructed, 0, nil, "Controls")
	for _, control := range controls {
		packet.AppendChild(control.Encode())
	}
	return packet
}
