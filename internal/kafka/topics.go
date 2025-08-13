package kafka

const (
	TopicEnqueued = "notifications.enqueued"

	TopicReadyEmailHigh = "notifications.ready.email.high"
	TopicReadyEmailLow  = "notifications.ready.email.low"
	TopicReadySMSHigh   = "notifications.ready.sms.high"
	TopicReadySMSLow    = "notifications.ready.sms.low"
	TopicReadyPushHigh  = "notifications.ready.push.high"
	TopicReadyPushLow   = "notifications.ready.push.low"
	TopicReadyInAppHigh = "notifications.ready.inapp.high"
	TopicReadyInAppLow  = "notifications.ready.inapp.low"
)