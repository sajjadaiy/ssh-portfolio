package main

type Item struct {
	Title       string
	Category    string
	Description string
	TechStack   string
	Tag         string
	Icon        string
	Link        string
	Repo        string // "owner/repo" for GitHub hydration; empty if N/A
	Year        string
	Role        string
}

type Social struct {
	Icon string
	Name string
	URL  string
	Link string
}

var items = []Item{
	{
		Title:    "FlowOps",
		Category: "projects",
		Tag:      "Infrastructure",
		Icon:     "◈",
		Link:     "https://gitlab.com/sajjadaiy/flowops",
		Repo:     "KingSajxxd/flowops",
		Year:     "2025",
		Role:     "Lead",
		TechStack: "WSO2 · Apache Kafka · CI/CD · Docker",
		Description: "Production-grade distributed systems environment simulating a real-world enterprise backend. " +
			"Engineered with a WSO2 API Gateway for secure routing and strict rate-limiting, " +
			"Apache Kafka for event-driven processing, and fully automated CI/CD pipelines.",
	},
	{
		Title:    "CACOP",
		Category: "projects",
		Tag:      "FinOps",
		Icon:     "⬡",
		Link:     "https://github.com/KingSajxxd/cacop",
		Repo:     "KingSajxxd/cacop",
		Year:     "2025",
		Role:     "Solo",
		TechStack: "Kubernetes · Chaos Mesh · FastAPI · Prometheus · Grafana · Loki",
		Description: "Cost-Aware Chaos Optimization Platform. A localized FinOps and chaos engineering " +
			"platform on Kubernetes. Uses Chaos Mesh for fault injection and a FastAPI control plane " +
			"with a full observability stack to simulate system failures and quantify compute waste.",
	},
	{
		Title:    "SSH Portfolio",
		Category: "projects",
		Tag:      "Systems",
		Icon:     "◉",
		Link:     "https://github.com/KingSajxxd/ssh-portfolio",
		Repo:     "KingSajxxd/ssh-portfolio",
		Year:     "2025",
		Role:     "Solo",
		TechStack: "Go · Wish · Bubble Tea · AWS EC2 · Docker",
		Description: "This terminal — accessible via ssh sajjad.tech. Built to study complex networking " +
			"architecture firsthand and create a developer-to-developer experience. Designed a secure " +
			"hybrid setup using reverse SSH tunneling through AWS EC2 to expose a home server behind " +
			"ISP restrictions, with container isolation and key-based authentication.",
	},
	{
		Title:    "Sajjad Aiyoob",
		Category: "about",
		Tag:      "Profile",
		Icon:     "●",
		TechStack: "Backend · DevOps · Cloud",
		Description: "Software Engineering Undergraduate at IIT/Westminster. " +
			"Focused on backend development and cloud-native systems. " +
			"I don't just write code — I ship systems.",
	},
}

var socials = []Social{
	{Icon: "◐", Name: "GitHub", URL: "github.com/KingSajxxd", Link: "https://github.com/KingSajxxd"},
	{Icon: "◧", Name: "LinkedIn", URL: "linkedin.com/in/sajjad-aiyoob", Link: "https://linkedin.com/in/sajjad-aiyoob"},
	{Icon: "✉", Name: "Email", URL: "sajaiyoobofficial@gmail.com", Link: "mailto:sajaiyoobofficial@gmail.com"},
}
