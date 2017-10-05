properties([buildDiscarder(logRotator(daysToKeepStr: '', numToKeepStr: '20'))])

notifyStarted()

node ("docker") {
    try {
        stage "checkout"
        checkout scm

        stage 'install'
        sh "docker-compose run --rm go make clean dependencies"

        stage 'test'
        sh "docker-compose run --rm go make test benchmark integration"

        stage 'build'
        sh "docker-compose run --rm  go ./ci/cross_compile.sh \"${env.BRANCH_NAME}\""

        stage 'post build'
        if (env.BRANCH_NAME == "master") {
            sh "make upload"
        }
        sh "docker-compose run --rm go make clean"

        notifySuccessful()
    } catch (e) {
      currentBuild.result = "FAILED"
      notifyFailed()
      throw e
    }
}

def notifyStarted() {
    slackSend (color: '#FFFF00', message: "STARTED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
}

def notifySuccessful() {
    slackSend (color: '#00FF00', message: "SUCCESSFUL: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
}

def notifyFailed() {
  slackSend (color: '#FF0000', message: "FAILED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
}
