# Challenge Statement

This challenge is about creating a simple video storage server with REST APIs

## Details

You are tasked to develop a simple video storage server with REST APIs, which should have:
- **CRUD implemention** as described in the [Open API definition](./api.yaml). (This document only contains the minimum and may need to be added).
- **Dockerfile** and **docker-compose** to build and run the server and other required services as docker containers.
- The endpoints of the server are exposed to the host machine.

## What we expect

When working on this challenge, be sure to:

- prove correctness of the code. We don't expect 100% test coverage but highlights on critical paths and logic is very welcome.
  
- handle errors and bad inputs
  
- provide user friendliness of installation and setup. We'll run `docker-compose up` in a clean environment without toolchains, JVM or SDKs and expect to see a server and the needed containers building and starting (this includes DB and all the other images used to complete the task).

We understand taking a challenge is time consuming, so feel free to choose an additional feature you feel passionate about and explain in words how you would like to implement it. We can discuss it further during the next interview steps!
See the [Bonus point and extensions](#bonus-points-and-extensions) section.
  

## How to submit your solution

- Push your code to this repository in the `main` branch.
- If you want to split the extension code from the main solution please submit it in a `ext-solution` branch
- Make sure the endpoints follow the path suggested in the `api.yaml` file (v1 included!).
- If your setup is correct the basic "health check" CI will return a green light and you can move forward. 

⚠️ **Note**: the CI/CD action tests only `v1/health` and it's not a guarantee of the entire solution correctness, but _without a green check we won't review the challenge_ as we can safely assume the overall solution is misconfigured.


### Bonus points and extensions

After implementing the above, feel free to implement one of the following features:

- Video searching by name, duration and so on

- Authentication and multi-user support

- Asynchronous video convert to webm (using ffmpeg etc.)

- Create an API that shows the progress of the converting task.

- Use same API to download converted data

- Data encryption and key management

*Note*

If you add or change APIs, include its OpenAPI document. However, please note that your server may be accessed by external clients in accordance with the given OpenAPI document and automated tests will hit the endpoints as described in [api.yaml](./api.yaml), therefore any change in the base path could result in 404 or false negative.
