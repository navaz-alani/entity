package eField

/*
These are tags which are used to provide specifications
for Entities and their fields.
*/
const (
	/*
		AxisTag is a tag used for tagging fields as
		axis fields for a particular struct.
	*/
	AxisTag string = "_ax_"
	/*
		IndexTag is used for tagging fields whose index
		needs to be created.
	*/
	IndexTag string = "_ix_"
	/*
		IDTag is used for providing Entity identifiers
		for an EntityMux.
	*/
	IDTag string = "_id_"
	/*
		HandleTag is used to provide configuration for
		generating pre-processing middleware.
	*/
	HandleTag string = "_hd_"
	/*
		RequireTag is used to tag struct fields which
		must be defined before a database entry.
	*/
	RequireTag string = "_rq_"
)
