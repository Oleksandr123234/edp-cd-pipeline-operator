# Architecture Scheme of CD Pipeline Operator

This page contains a representation of the current CD Pipeline Operator architecture that is built using the plantUML capabilities.
All the diagrams sources are placed under the **/puml** directory of the current folder.

An Image of the HEAD of the current branch will be displayed as a result of an Image building with the plant uml proxy server.

If you are in the detached mode, use the sources to get the required version of diagrams.

Since EDP v3.3.0, the automatic deployment feature has been added for the Tekton deploy scenario. Under the hood, the auto-deploy logic for the Tekton deploy scenario is implemented this way:

  1. User clicks the **Build** button or merges patch to VCS.
  2. If the build is successful, new tag is appended to the `CodebaseImageStream` resource.
  3. The `codebase-operator` spots the new tag and creates the `CDStageDeploy` with this tag.
  4. The `codebase-operator` takes the new tag out of the `CDStageDeploy` resource and updates image tag in Argo CD.
  5. Lastly, Argo CD deploys newer image.

In Jenkins deploy scenario, the autodeploy feature is implemented a little more complicated but still remains similar:

  1. Once the stage with the enabled autodeploy feature is created, CD pipeline processes this stage and creates corresponding Jenkins job with the `autodeploy: true` parameter.
  2. User clicks the **Build** button or merges patch to VCS.
  2. When application build is launched, Jenkins attaches specific tag to the CodebaseImageStream. This tag is further processed by the `codebase-operator`. As a result, the `CDStageDeploy` resource is created at the end of the processing.
  3. Further the `CDStageDeploy` resource is processed by the `codebase-operator`. The `CDStageJenkinsDeployment` is created at the end of the process.
  4. Finally, the `jenkins-operator` processes the `CDStageJenkinsDeployment` resource and triggers the Jenkins deploy job.

![arch](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/epam/edp-cd-pipeline-operator/master/docs/puml/arch.puml)
