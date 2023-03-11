import { Construct } from "constructs";
import { App, AssetType, TerraformAsset, TerraformStack } from "cdktf";
import * as google from '@cdktf/provider-google';
import * as path from 'path';

const project = 'web-sakura';
const region = 'us-central1';
const repository = 'web-sakura';

class MyStack extends TerraformStack {
  constructor(scope: Construct, id: string) {
    super(scope, id);

    new google.provider.GoogleProvider(this, 'google', {
      project,
    });

    const autoRegistRunner = new google.serviceAccount.ServiceAccount(this, 'autoRegistRunner', {
      accountId: 'auto-regist-runner',
    });

    new google.projectIamMember.ProjectIamMember(this, 'allowSecretAccess', {
      member: `serviceAccount:${autoRegistRunner.email}`,
      project,
      role: 'roles/secretmanager.secretAccessor',
    });

    const accountSecret = new google.secretManagerSecret.SecretManagerSecret(this, 'account', {
      secretId: 'account',
      replication: {
        automatic: true,
      },
    });

    const passwordSecret = new google.secretManagerSecret.SecretManagerSecret(this, 'password', {
      secretId: 'password',
      replication: {
        automatic: true,
      },
    });

    const childIDSecret = new google.secretManagerSecret.SecretManagerSecret(this, 'childID', {
      secretId: 'child-id',
      replication: {
        automatic: true,
      },
    });

    const asset = new TerraformAsset(this, 'asset', {
      path: path.resolve('auto-regist'),
      type: AssetType.ARCHIVE,
    });    

    const assetBucket = new google.storageBucket.StorageBucket(this, 'assetBucket', {
      lifecycleRule: [{
        action: {
            type: 'Delete',
        },
        condition: {
            age: 1,
        },
      }],
      location: region,
      name: `${project}-asset`,
    });

    const assetObject = new google.storageBucketObject.StorageBucketObject(this, 'assetObject', {
      bucket: assetBucket.name,
      name: asset.assetHash,
      source: asset.path,
    });

    const autoRegist = new google.cloudfunctionsFunction.CloudfunctionsFunction(this, 'autoRegist', {
        entryPoint: 'EntryPoint',
        ingressSettings: 'ALLOW_ALL',
        minInstances: 0,
        maxInstances: 1,
        name: 'auto-regist',
        region,
        runtime: 'go120',
        secretEnvironmentVariables: [
            {
                key: 'ACCOUNT',
                projectId: project,
                secret: accountSecret.secretId,
                version: '1',
              },
              {
                key: 'PASSWORD',
                projectId: project,
                secret: passwordSecret.secretId,
                version: '1',
              },
              {
                key: 'CHILD_ID',
                projectId: project,
                secret: childIDSecret.secretId,
                version: '1',
              },      
        ],
        serviceAccountEmail: autoRegistRunner.email,
        sourceArchiveBucket: assetBucket.name,
        sourceArchiveObject: assetObject.name,
        triggerHttp: true,
    });

    new google.cloudSchedulerJob.CloudSchedulerJob(this, 'schedule', {
      name: 'auto-regist-schedule',
      httpTarget: {
        uri: autoRegist.httpsTriggerUrl,
      },
      region,
      schedule: '0 0 * * *',
    });

    new google.cloudbuildTrigger.CloudbuildTrigger(this, 'buildTrigger', {
      filename: 'cloudbuild.yaml',
      github: {
        owner: 'hsmtkk',
        name: repository,
        push: {
          branch: 'main',
        },
      },
    });
  }
}

const app = new App();
new MyStack(app, "web-sakura");
app.synth();
