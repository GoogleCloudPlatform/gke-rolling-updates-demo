#!/usr/bin/env groovy
/*
Copyright 2019 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

// The declarative agent is defined in yaml.  It was previously possible to
// define containerTemplate but that has been deprecated in favor of the yaml
// format
// Reference: https://github.com/jenkinsci/kubernetes-plugin

// set up pod label and GOOGLE_APPLICATION_CREDENTIALS (for Terraform)
pipeline {
  agent {
    kubernetes {
      label 'k8s-infra-rolling-upgrades'
      yaml """
apiVersion: v1
kind: Pod
metadata:
  labels:
    jenkins: build-node
spec:
  containers:
  - name: k8s-node-expand-contract
    image: gcr.io/pso-helmsman-cicd/jenkins-k8s-node:${env.CONTAINER_VERSION}
    command: ['cat']
    tty: true
    env:
    - name: KUBECONFIG
      value: /home/jenkins/.kube/expand-contract-config
    volumeMounts:
    # Mount the dev service account key
    - name: dev-key
      mountPath: /home/jenkins/dev
  - name: k8s-node-rolling-upgrade
    image: gcr.io/pso-helmsman-cicd/jenkins-k8s-node:${env.CONTAINER_VERSION}
    command: ['cat']
    tty: true
    env:
    - name: KUBECONFIG
      value: /home/jenkins/.kube/rolling-upgrade-config
    volumeMounts:
    # Mount the dev service account key
    - name: dev-key
      mountPath: /home/jenkins/dev
  - name: k8s-node-blue-green
    image: gcr.io/pso-helmsman-cicd/jenkins-k8s-node:${env.CONTAINER_VERSION}
    command: ['cat']
    tty: true
    env:
    - name: KUBECONFIG
      value: /home/jenkins/.kube/blue-green-config
    volumeMounts:
    # Mount the dev service account key
    - name: dev-key
      mountPath: /home/jenkins/dev
  volumes:
  # Create a volume that contains the dev json key that was saved as a secret
  - name: dev-key
    secret:
      secretName: jenkins-deploy-dev-infra
"""
    }
  }
  environment {
    GOOGLE_APPLICATION_CREDENTIALS = '/home/jenkins/dev/jenkins-deploy-dev-infra.json'
  }
  stages {
    stage('Setup') {
      steps {
        container('k8s-node-expand-contract') {
          // checkout code from scm i.e. commits related to the PR
          checkout scm

          // Setup gcloud service account access
          sh "gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}"
          sh "gcloud config set compute/zone ${env.CLUSTER_ZONE}"
          sh "gcloud config set core/project ${env.PROJECT_ID}"
          sh "gcloud config set compute/region ${env.REGION}"
        }
        container('k8s-node-rolling-upgrade') {
          // checkout code from scm i.e. commits related to the PR
          checkout scm

          // Setup gcloud service account access
          sh "gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}"
          sh "gcloud config set compute/zone ${env.CLUSTER_ZONE}"
          sh "gcloud config set core/project ${env.PROJECT_ID}"
          sh "gcloud config set compute/region ${env.REGION}"
        }
        container('k8s-node-blue-green') {
          // checkout code from scm i.e. commits related to the PR
          checkout scm

          // Setup gcloud service account access
          sh "gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}"
          sh "gcloud config set compute/zone ${env.CLUSTER_ZONE}"
          sh "gcloud config set core/project ${env.PROJECT_ID}"
          sh "gcloud config set compute/region ${env.REGION}"
        }
      }
    }
    stage('Configure environment') {
      steps {
        container('k8s-node-expand-contract') {
          sh './test/configure-jenkins-environment.sh'
        }
        container('k8s-node-rolling-upgrade') {
          sh './test/configure-jenkins-environment.sh'
        }
        container('k8s-node-blue-green') {
          sh './test/configure-jenkins-environment.sh'
        }
      }
    }
    stage('Lint') {
      steps {
        container('k8s-node-expand-contract') {
          sh "make lint"
        }
      }
    }
    stage('Run Tests') {
      parallel {
        stage('Expand/Contract Upgrade') {
          steps {
            container('k8s-node-expand-contract') {
              sh 'make expand-contract-upgrade'
            }
          }
          post {
            always {
              container('k8s-node-expand-contract') {
                sh 'make expand-contract-upgrade-delete'
              }
            }
          }
        }
        stage('In-Place Rolling Upgrade') {
          steps {
            container('k8s-node-rolling-upgrade') {
              sh 'make in-place-rolling-upgrade'
            }
          }
          post {
            always {
              container('k8s-node-rolling-upgrade') {
                sh 'make in-place-rolling-upgrade-delete'
              }
            }
          }
        }
        stage('Blue/Green Upgrade') {
          steps {
            container('k8s-node-blue-green') {
              sh 'make blue-green-upgrade'
            }
          }
          post {
            always {
              container('k8s-node-blue-green') {
                sh 'make blue-green-upgrade-delete'
              }
            }
          }
        }
      }
    }
  }
}
