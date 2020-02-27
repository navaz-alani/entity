/*
Package multiplexer defines the EMux type which is basically
a multiplexer for Entity types.

The main goal is to enable users to be more productive by writing
less repetitive boilerplate code for each entity. This is particularly
true when writing API handlers for CRUD operations. For example,
consider registering a User into the database. The request handler for
this purpose would decode a JSON payload from the request and attempt
to create and save a User entity instance, which will then be added to
the database. When there are many such entities managed within an app,
this can become repetitive and make larger changes harder to implement.

The multiplexer is able to handle such cases by providing middleware
for pre-processing requests. For example, once the multiplexer has parsed
schema definitions, the CreationMiddleware method can be used to generate
middleware to pre-process incoming JSON payloads directly into the Entity
intended. This way, the user only has to care about WHAT to do with the
Entity.

Middleware generation is planned for the four CRUD operations: creation
(as described above), retrieval, update and delete operations. This feature
has not been fully completed but has been planned for implementation.

Entity definitions are expected to be of kind "struct". This package also
makes use of a set of tags which can be used to provide Entity specifications
to the multiplexer. These tags are discussed below.

Tags

Here are the eField tags that the EMux uses:

entity.IDTag - This tag is used to give a name to an Entity.
This name specifies the mongo.Collection that will be created
in the database for an Entity. It is also used by EMux to
internally work with Entity types. This value must be unique
amongst the Entity types that the EMux manages.

entity.HandleTag - This tag is used to provide configurations
for middleware generation. The value for this tag is a string
containing configuration tokens. These tokens are single characters
(runes) which can be used to classify an eField. For example, the
CreationFieldsToken token can be used used to specify which
fields should be parsed from an http.Response body for the
middleware generation.

entity.AxisTag - This tag is used to specify which fields can be
considered to be unique (to an Entity) within a collection.
The tag value which indicates that an eField is an axis eField is
the string "true"-- all other values are rejected.

entity.IndexTag - This tag is used to specify the fields for which
an index needs to be built in the database collection. This is used
hand in hand with the entity.Axis tag; in order for a eField's index
to be constructed, both these tags have to be set to "true".
*/
package multiplexer
