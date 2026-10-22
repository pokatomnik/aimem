package fact

type (
	FactListView struct {
		ID        string
		CreatedAt string
		Summary   string
	}

	FactFullView struct {
		ID        string
		CreatedAt string
		Summary   string
		Content   string
	}
)
