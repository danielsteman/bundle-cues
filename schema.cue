#Job: {
    name: string
    ...
}

#Channel: {
    id: string
}

#Schema: {
    include: [...string]
    resources: {
        jobs: {
            [string]: #Job
        }
        ...
    }
    // "a"!: string
    targets: {
        prod: {
            resources: {
                jobs!: {
                    [string]: {
                        webhook_notifications!: {
                            on_failure: {
                                [...#Channel]
                            }
                        }
                        ...
                    }
                }
                ...
            }
            ...
        }
        ...
    }
    ...
}

#Schema
