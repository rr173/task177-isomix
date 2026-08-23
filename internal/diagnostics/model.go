package diagnostics

// Report describes numerical and structural properties of one solution.
type Report struct {
	SolutionID     string   `json:"solution_id"`
	Status         string   `json:"status"`
	ConstraintRows int      `json:"constraint_rows"`
	Variables      int      `json:"variables"`
	Density        float64  `json:"density"`
	Residual       float64  `json:"residual"`
	Risk           string   `json:"risk"`
	Observations   []string `json:"observations,omitempty"`
}
