package client

// AWX API response structures
type JobTemplate struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type JobTemplatesResponse struct {
	Count   int           `json:"count"`
	Results []JobTemplate `json:"results"`
}

type Job struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

type JobsResponse struct {
	Count   int   `json:"count"`
	Results []Job `json:"results"`
}

type JobLaunchRequest struct {
	ExtraVars map[string]interface{} `json:"extra_vars,omitempty"`
}

type JobLaunchResponse struct {
	Job           int                    `json:"job"`
	IgnoredFields map[string]interface{} `json:"ignored_fields,omitempty"`
	ID            int                    `json:"id"`
	Type          string                 `json:"type"`
	URL           string                 `json:"url"`
}
