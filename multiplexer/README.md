# Multiplexer Internals

This is high level overview of the internals of the multiplexer.
The following structure is similar to a definition in Go. The angle
bracket tags represent the internal type (type defined and used in
the source code). For example, `<EMux>` declares that the next 
definition is for the internal type `EMux`. Similarly, `<ptr_Entity>`
declares that the next definition is for the internally used type
`*Entity`.


```xml
<EMux>struct{
    TypeMap  <TypeMap>map[reflect.Type]string
	Entities <EntityMap>map[string]<ptr_metaEntity>*struct{
		EntityID             string
		Entity   <ptr_Entity>*struct{
			SchemaDefinition reflect.Type
			PStorage        *mongo.Collection
		},
		FieldClassifications map[rune][]<ptr_condensedField>*struct{
			Name      string
			Type      reflect.Type
			RequestID string
			Value     interface{}
			Embedded  <Embedding>*struct{
				CFlag bool
				CType reflect.Type
				SFlag bool
				SType reflect.Type
				Meta  <ptr_metaEntity>
			},
		},	
	},
}
```

## Quick Overview

This section provides a short description of the function and intent of
every type used in the above structure.

The `EMux` is the multiplexer itself. It contains 2 fields. An `Entities` 
field, which contains a map of strings (which are the EntityIDs) to 
`<ptr_metaEntity>`. And an `TypeMap` is the inverse of the `EntityMap`. This 
one maps Entity types (reflect.Type) to a string containing the EntityID of 
the Entity.

The following narrows down to the `Entities` field specifically.

The `metaEntity` type is used to package information about an Entity being
managed by the multiplexer. It provides access to pre-computed values to be
used for the multiplexer's operations. 

Quickly, the EntityID field is a string declaring an Entity's EntityID. Entity 
is a field which contains a pointer to an `Entity` type containing the type
specifications (reflect.Type) of the Entity, as well as a pointer to the 
mongoDB collection being used for this Entity (`nil` if no collection).

The `FieldClassifications` field is a map mapping a character to a slice of 
pointers to a `condensedField`s. This is a short-hand representation of a 
field, also containing specific information needed by the multiplexer for its 
operations. The characters (runes) that map to these pointers are reserved 
tokens which represent a specific classification of a field. For example, 
currently, the rune `'*'` is a token which maps to a pointer to the
condensedField which provided the EntityID being used.  

`Name` is the field's name and `Type` is the field's type. `RequestID` is a
string indicating the field's name in JSON payloads. `Value` stores additional
information in special cases (e.g. EntityID).

The `Embedded` field conatins an `Embedding` type instance which indicates the
nature of nesting for the field. `CFlag` and `SFlag` indicate whether the field
stores a collection and a struct respectively. The Type analogs of these
specify the embedded type and `Meta` stores a pointer to the `metaEntity` for
that type.
