# process_async_ds

![docker-image](https://github.com/JLCodeSource/process_async_ds/actions/workflows/docker-image.yml/badge.svg)

![go-linters](https://github.com/JLCodeSource/process_async_ds/actions/workflows/go-linters.yml/badge.svg)

Process Async DS is a tool that manages processed data files in object storage.
It uses metadata to identify files that have completed a specific processing step and then moves them to a new local location for further processing.
This automated movement streamlines workflows and prepares the data for subsequent operations.