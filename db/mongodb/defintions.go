package mongodb

import "context"

const (
	// MULTIPLELike used only for host search
	MULTIPLELike = "$multilike"

	// IN the db operator
	IN = "$in"

	// OR the db operator
	OR = "$or"

	// AND the db operator
	AND = "$and"

	// LIKE the db operator
	LIKE = "$regex"

	// OPTIONS the db operator,used with $regex
	// detail to see https://docs.mongodb.com/manual/reference/operator/query/regex/#op._S_options
	OPTIONS = "$options"

	// EQ the db operator
	EQ = "$eq"

	// NE the db operator
	NE = "$ne"

	// NIN the db oeprator
	NIN = "$nin"

	// LT the db operator
	LT = "$lt"

	// LTE the db operator
	LTE = "$lte"

	// GT the db operator
	GT = "$gt"

	// GTE the db opeartor
	GTE = "$gte"

	// Exists the db opeartor
	Exists = "$exists"

	// Not the db opeartor
	Not = "$not"

	// Count the db opeartor
	Count = "$count"

	// Group the db opeartor
	Group = "$group"

	// Match the db opeartor
	Match = "$match"

	// Sum the db opeartor
	Sum = "$sum"

	// Push the db opeartor
	Push = "$push"

	// UNSET the db opeartor
	UNSET = "$unset"

	// AddToSet The $addToSet operator adds a value to an array unless the value is already present, in which case $addToSet does nothing to that array.
	AddToSet = "$addToSet"

	// Pull The $pull operator removes from an existing array all instances of a value or values that match a specified condition.
	Pull = "$pull"

	// All matches arrays that contain all elements specified in the query.
	All = "$all"

	// Project passes along the documents with the requested fields to the next stage in the pipeline
	Project = "$project"

	// Size counts and returns the total number of items in an array
	Size = "$size"

	// Type selects documents where the value of the field is an instance of the specified BSON type(s).
	// Querying by data type is useful when dealing with highly unstructured data where data types are not predictable.
	Type = "$type"

	// Sort the db operator
	Sort = "$sort"

	// ReplaceRoot the db operator
	ReplaceRoot = "$replaceRoot"

	// Limit the db operator to limit return number of doc
	Limit = "$limit"

	// Unwind used to split values contained in an array field into separate doc
	Unwind = "$unwind"

	// Lookup used to perform multi-table association operations
	Lookup = "$lookup"

	GraphLookup = "$graphLookup"

	// From collection to join
	From = "from"

	// LocalField field from the input documents
	LocalField = "localField"

	// ForeignField field from the documents of the "from" collection
	ForeignField = "foreignField"

	// As output array field
	As = "as"

	// Skip skip data index
	Skip = "$skip"

	//// BKLimit data quantity limit
	//Limit = "$limit"
	// AddFields add pipeline field
	AddFields = "$addFields"
	// Convert filed type
	Convert = "$convert"

	// ToString
	ToString = "$toString"
)

const (
	FieldMongoID   = "_id"
	FieldID        = "id"
	FieldName      = "name"
	DefaultField   = "default"
	FieldUserID    = "user_id"
	FieldDeletedAt = "deleted_at"
)

type ReadPreferenceMode string

// String 用于打印
func (r ReadPreferenceMode) String() string {
	return string(r)
}

const (
	// NilMode not set
	NilMode ReadPreferenceMode = ""
	// PrimaryMode indicates that only a primary is
	// considered for reading. This is the default
	// mode.
	PrimaryMode ReadPreferenceMode = "1"
	// PrimaryPreferredMode indicates that if a primary
	// is available, use it; otherwise, eligible
	// secondaries will be considered.
	PrimaryPreferredMode ReadPreferenceMode = "2"
	// SecondaryMode indicates that only secondaries
	// should be considered.
	SecondaryMode ReadPreferenceMode = "3"
	// SecondaryPreferredMode indicates that only secondaries
	// should be considered when one is available. If none
	// are available, then a primary will be considered.
	SecondaryPreferredMode ReadPreferenceMode = "4"
	// NearestMode indicates that all primaries and secondaries
	// will be considered.
	NearestMode ReadPreferenceMode = "5"
)

func GetDBReadPreference(ctx context.Context) ReadPreferenceMode {
	val := ctx.Value(HTTPReadReference)
	if val != nil {
		mode, ok := val.(string)
		if ok {
			return ReadPreferenceMode(mode)
		}
	}
	return NilMode
}

const (
	HTTPReadReference = "Read_Preference"
)
