/*
Package entity defines a convenient abstraction which
can be used with MongoDB in order to streamline CRUD
operations.

This definition also allows a centralization of security
and other policies related to an entity within the app.
This leads to fewer bugs, and when there are bugs, due
to the modularization of policies, the origin of bugs
is easier to locate.

The goal is to define an abstraction which is useful for
the general entity and can make the process of writing code
much more efficient.

Axis Policy

An axis is defined as a eField in an Entity which
can be assumed to be unique. This is important when creating
collection indexes and creating Query/Update/Delete filters.
The Axis Policy ensures data integrity by enforcing that all
Entities within a collection have unique axis values.

This Policy is especially useful when querying elements for
Read/Update/Delete operations. The client benefits through the
ability to specify whether a eField is an axis, using the "axis"
tag, in the struct eField. This tag can be set to "true" to enforce
it as an axis eField.

Getting started

To use the Entity abstraction, start by creating a struct
which will define the Entity that you want to work with.
Available in this step, is the "axis" tag which is useful in
specifying which fields are to be treated as axis fields.
For example, here is a hypothetical struct for defining
a (useless) User Entity:

	type User struct {
		ID    primitive.ObjectID `_id_:"user" json:"-" bson:"_id"`
		Name  string             `json:"name" _hd_:"c"`
		Email string             `json:"email" axis:"true" index:"text" _hd_:"c"`
	}

Next, register this User struct as an Entity.

	UserEntity := Entity{
		SchemaDefinition: TypeOf(User{}),
		PStorage:         &mongoCollection
	}

Run the Optimize function to generate indexes for the axis fields:

	UserEntity.Optimize()

Create a User:

	u := User{
		Name:  "Jane Doe",
		Email: "jane.doe@example.com"
	}

Add this user to the database:

	id, err := UserEntity.Add(u)

The other Read/Update/Delete operations become as simple with
this Entity definition and can lead to hundreds of lines of less
code being written.
The Entity package aims to (at least) provide a high level
API with the basic CRUD boilerplate code already taken care of.

See github.com/navaz-alani/entity/multiplexer for information about
the EntityMux which is able to manage a collection of entities for
larger applications.
*/
package entity
