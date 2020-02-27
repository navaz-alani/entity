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
    TypeMap  <TypeMap>map[reflect.Type]string,
	Entities <EntityMap>map[string]<ptr_metaEntity>*struct{
		EntityID             string,
		Entity   <ptr_Entity>*struct{
			SchemaDefinition reflect.Type,
			PStorage        *mongo.Collection,
		},
		FieldClassifications map[rune][]<ptr_condensedField>*struct{
			Name      string,
			Type      reflect.Type,
			RequestID string,
			Value     interface{},
			Embedded  <ptr_metaEntity>,
		},	
	},
}
```

## Quick Overview

This section provides a short description of the function and intent of
every type used in the above structure.

The `EMux` is the multiplexer itself. It contains 2 fields. An `Entities` field,
which contains a map of strings (which are the EntityIDs) to `<ptr_metaEntity>`.
The `metaEntity` type is used to package information about an Entity being
managed by the multiplexer. It provides access to pre-computed values to be
used for the multiplexer's operations. 

Quickly, the EntityID field is a string
declaring an Entity's EntityID. Entity is a field which contains a pointer to
an `Entity` type containing the type specifications (reflect.Type) of the Entity,
as well as a pointer to the `mongoDB` collection being used for this Entity (if 
any).

The `FieldClassifications` field is a map mapping a character to an array of pointers
 to a `condensedField`s. This is a short-hand representation of a field, also 
containing specific information needed by the multiplexer for its operations.
The characters (runes) that map to these pointers are reserved tokens which
represent a specific classification of a field. For example, currently, the rune `'*'` is a token which maps to a pointer to the condensedField which provided
the EntityID being used.  

Finally, the `TypeMap` is the inverse of the `EntityMap`. This one maps Entity
types (reflect.Type) to a string containing the EntityID of the Entity.