// ─────────────────────────────────────────────────────────────
// Jenkinsfile – Notification & Messaging Service Pipeline
// Penanggung Jawab: Muhammad Fittra Novria Rizky
// ─────────────────────────────────────────────────────────────
pipeline {
    agent any

    environment {
        SERVICE_NAME    = "notification-service"
        IMAGE_NAME      = "notification-service"
        IMAGE_TAG       = "${env.BUILD_NUMBER}-${env.GIT_COMMIT?.take(7) ?: 'dev'}"
        REGISTRY        = "registry.internal.id"
        K8S_NAMESPACE   = "messaging"
        // DB DSN untuk functional test (injected oleh Jenkins Credentials)
        FUNCTIONAL_TEST_DB_DSN = credentials('notification-service-func-test-db-dsn')
    }

    options {
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '10'))
    }

    stages {
        // ──────────────────────────────────────────
        // 1. CHECKOUT
        // ──────────────────────────────────────────
        stage('Checkout') {
            steps {
                checkout scm
                echo "Branch   : ${env.BRANCH_NAME}"
                echo "Commit   : ${env.GIT_COMMIT}"
                echo "Build #  : ${env.BUILD_NUMBER}"
            }
        }

        // ──────────────────────────────────────────
        // 2. UNIT TESTS  (tidak boleh akses DB)
        // ──────────────────────────────────────────
        stage('Unit Tests') {
            steps {
                sh '''
                    echo "=== Running Unit Tests ==="
                    go test ./internal/... \
                        -v \
                        -count=1 \
                        -race \
                        -coverprofile=coverage_unit.out \
                        -covermode=atomic \
                        -timeout 60s
                '''
            }
            post {
                always {
                    sh 'go tool cover -func=coverage_unit.out || true'
                    publishHTML(target: [
                        reportDir  : '.',
                        reportFiles: 'coverage_unit.html',
                        reportName : 'Unit Test Coverage'
                    ])
                }
            }
        }

        // ──────────────────────────────────────────
        // 3. LINT / VET
        // ──────────────────────────────────────────
        stage('Lint / Vet') {
            steps {
                sh '''
                    echo "=== Running go vet ==="
                    go vet ./...

                    echo "=== Running staticcheck (opsional) ==="
                    which staticcheck && staticcheck ./... || echo "staticcheck tidak tersedia, skip"
                '''
            }
        }

        // ──────────────────────────────────────────
        // 4. BUILD IMAGE (lokal, belum di-push)
        // ──────────────────────────────────────────
        stage('Build Image') {
            steps {
                sh '''
                    echo "=== Building Docker Image ==="
                    docker build \
                        --build-arg BUILD_NUMBER=${BUILD_NUMBER} \
                        --build-arg GIT_COMMIT=${GIT_COMMIT} \
                        -t ${IMAGE_NAME}:${IMAGE_TAG} \
                        -t ${IMAGE_NAME}:latest \
                        .
                '''
            }
        }

        // ──────────────────────────────────────────
        // 5. FUNCTIONAL TESTS  (akses DB diizinkan)
        // ──────────────────────────────────────────
        stage('Functional Tests') {
            environment {
                // Override DSN agar mengarah ke test DB yang sudah running di lokal/staging
                FUNCTIONAL_TEST_DB_DSN = "${FUNCTIONAL_TEST_DB_DSN}"
            }
            steps {
                sh '''
                    echo "=== Starting test dependencies (docker-compose) ==="
                    docker-compose -f docker-compose.test.yml up -d postgres rabbitmq
                    sleep 5  # tunggu DB ready

                    echo "=== Running Functional Tests ==="
                    go test ./test/functional/... \
                        -v \
                        -tags=functional \
                        -count=1 \
                        -timeout 120s

                    echo "=== Stopping test dependencies ==="
                    docker-compose -f docker-compose.test.yml down
                '''
            }
            post {
                always {
                    sh 'docker-compose -f docker-compose.test.yml down --volumes || true'
                }
            }
        }

        // ──────────────────────────────────────────
        // 6. PUSH IMAGE ke Registry
        // ──────────────────────────────────────────
        stage('Push Image') {
            when {
                anyOf {
                    branch 'main'
                    branch 'release/*'
                }
            }
            steps {
                withCredentials([usernamePassword(
                    credentialsId: 'registry-credentials',
                    usernameVariable: 'REG_USER',
                    passwordVariable: 'REG_PASS'
                )]) {
                    sh '''
                        echo "=== Pushing Docker Image ==="
                        echo ${REG_PASS} | docker login ${REGISTRY} -u ${REG_USER} --password-stdin
                        docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}
                        docker tag ${IMAGE_NAME}:latest ${REGISTRY}/${IMAGE_NAME}:latest
                        docker push ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}
                        docker push ${REGISTRY}/${IMAGE_NAME}:latest
                    '''
                }
            }
        }

        // ──────────────────────────────────────────
        // 7. DEPLOY di Kubernetes
        // ──────────────────────────────────────────
        stage('Deploy') {
            when {
                anyOf {
                    branch 'main'
                    branch 'release/*'
                }
            }
            steps {
                withKubeConfig([credentialsId: 'k8s-kubeconfig']) {
                    sh '''
                        echo "=== Deploying to Kubernetes ==="
                        kubectl set image deployment/${SERVICE_NAME} \
                            ${SERVICE_NAME}=${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG} \
                            -n ${K8S_NAMESPACE}

                        kubectl rollout status deployment/${SERVICE_NAME} \
                            -n ${K8S_NAMESPACE} \
                            --timeout=300s
                    '''
                }
            }
        }

        // ──────────────────────────────────────────
        // 8. VERIFY (smoke test post-deploy)
        // ──────────────────────────────────────────
        stage('Verify') {
            when {
                anyOf {
                    branch 'main'
                    branch 'release/*'
                }
            }
            steps {
                withKubeConfig([credentialsId: 'k8s-kubeconfig']) {
                    sh '''
                        echo "=== Verifying deployment ==="
                        # Cek pod running
                        kubectl get pods -n ${K8S_NAMESPACE} -l app=${SERVICE_NAME}

                        # Health-check endpoint
                        SERVICE_HOST=$(kubectl get svc ${SERVICE_NAME} -n ${K8S_NAMESPACE} \
                            -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
                        curl --fail --retry 5 --retry-delay 5 \
                            http://${SERVICE_HOST}:8080/health || exit 1

                        echo "=== Deployment verified successfully ==="
                    '''
                }
            }
        }
    }

    post {
        success {
            echo "Pipeline berhasil: ${SERVICE_NAME} build ${IMAGE_TAG}"
        }
        failure {
            echo "Pipeline GAGAL: cek log di atas"
        }
        always {
            cleanWs()
        }
    }
}
