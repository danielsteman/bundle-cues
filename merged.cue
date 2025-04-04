include: [
	"test_include.yml",
	"resources/*.yml",
]
targets: {
	dev: null
	prod: resources: jobs: some_job: webhook_notifications: on_failure: [{id: "suuuup"}]
}
resources: {
	jobs: {
		job123: name: "job123"
		some_job: name: "yooooo"
	}
	pipelines: ["sup"]
}
