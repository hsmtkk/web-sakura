import { Construct } from "constructs";
import { App, AssetType, TerraformAsset, TerraformStack } from "cdktf";
import * as google from '@cdktf/provider-google';
import * as path from 'path';

const project = 'web-sakura';
const region = 'us-central1';

class MyStack extends TerraformStack {
  constructor(scope: Construct, id: string) {
    super(scope, id);

    new google.provider.GoogleProvider(this, 'google', {
      project,
    });

    const autoRegistRunner = new google.serviceAccount.ServiceAccount(this, 'autoRegistRunner', {
      accountId: 'auto-regist-runner',
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
      location: region,
      name: `${project}-asset`,
    });

    const assetObject = new google.storageBucketObject.StorageBucketObject(this, 'assetObject', {
      bucket: assetBucket.name,
      name: asset.assetHash,
      source: asset.path,
    });

    const schedulerTopic = new google.pubsubTopic.PubsubTopic(this, 'schedulerTopic', {
      name: 'scheduler-topic',
    });

    new google.cloudfunctions2Function.Cloudfunctions2Function(this, 'autoRegist', {
      buildConfig: {
        entryPoint: 'EntryPoint',
        runtime: 'go120',
        source: {
          storageSource: {
            bucket: assetBucket.name,
            object: assetObject.name,
          },
        },
      },
      eventTrigger: {
        eventType: 'google.cloud.pubsub.topic.v1.messagePublished',
        pubsubTopic: schedulerTopic.id,
      },
      location: region,
      name: 'auto-regist',
      serviceConfig: {
        minInstanceCount: 0,
        maxInstanceCount: 1,
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
      },
    });

    new google.cloudSchedulerJob.CloudSchedulerJob(this, 'schedule', {
      name: 'auto-regist-schedule',
      pubsubTarget: {
        topicName: schedulerTopic.id,
      },
      region,
      schedule: '0 0 * * *',
    });
  }
}

const app = new App();
new MyStack(app, "web-sakura");
app.synth();
