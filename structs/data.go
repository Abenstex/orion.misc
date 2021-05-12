package structs

import (
	"github.com/abenstex/orion.commons/structs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StateTransitionRule struct {
	ID                  *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Info                structs.BaseInfo    `bson:"info" json:"info"`
	SourceState         string              `bson:"source_state" json:"source_state"`
	AllowedTargetStates []string            `bson:"allowed_target_states" json:"allowed_target_states"`
}

type AttributeChange struct {
	AttributeId   uint64
	ObjectType    string
	ObjectId      uint64
	ObjectVersion int
	OriginalValue string
	NewValue      string
}

type Hierarchy struct {
	ID      *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Info    structs.BaseInfo    `bson:"info" json:"info"`
	Entries []HierarchyEntry    `bson:"entries" json:"entries"`
}

func NewHierarchy() Hierarchy {
	hierarchy := Hierarchy{Entries: make([]HierarchyEntry, 0)}

	return hierarchy
}

type HierarchyEntry struct {
	ID         *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Index      int                 `bson:"index" json:"index"`
	ObjectType string              `bson:"object_type" json:"object_type"`
}

type Parameter struct {
	ID    *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Info  structs.BaseInfo    `bson:"info" json:"info"`
	Value string              `bson:"value" json:"value"`
}

type Category struct {
	ID             *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Info           structs.BaseInfo    `bson:"info" json:"info"`
	ReferencedType string              `bson:"referenced_type" json:"referenced_type"`
}

type ObjectTypeCustomization struct {
	ID                *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ObjectType        string              `bson:"object_type" json:"object_type"`
	FieldName         string              `bson:"field_name" json:"field_name"`
	FielDataType      string              `bson:"field_data_type" json:"field_data_type"`
	FieldMandatory    bool                `bson:"is_mandatory_field" json:"is_mandatory_field"`
	FieldDefaultValue *string             `bson:"field_default_value" json:"field_default_value"`
	CreatedDate       int64               `bson:"created_date" json:"created_date"`
	CreatedBy         *string             `bson:"created_by" json:"created_by"`
	UserComment       *string             `bson:"user_comment" json:"user_comment"`
	User              *string             `bson:"user" json:"user"`
	ChangeDate        *int64              `bson:"change_date" json:"change_date"`
}
