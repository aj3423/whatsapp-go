package wam

import (
	"math"

	"arand"
)

// jeb: "messageIsInternational"
// then search in file found:
//   for class: "case 1942: {"

var MapStats map[int32]*ClassAttr

type ClassAttr struct {
	Id     int32
	Desc   string
	Weight int // 3rd param of SamplingRate(x, y, weight)
	Stats  map[int32]string
}

func (c *ClassAttr) RateHit() bool {
	abs := int(math.Abs(float64(c.Weight)))
	return arand.Int(0, abs) == 0
}

var Wild = &ClassAttr{0, "Wild", 1, map[int32]string{
	47: `Timestamp`,

	289:  `BUILD_MODEL`,
	3:    `mnc`,
	5:    `mcc`,
	5029: `ab_props:sys:last_exposure_keys`,
	7335: `md-check-msgstore-not-ready`,
	105:  `NetworkSubType`,
	11:   `Fixed_11`,
	6251: `PkgSignCorrect`,
	2795: `NoiseLocation`,
	13:   `MANUFACTURER_MODEL`,
	15:   `BUILD_VERSION_RELEASE`,
	655:  `MemoryClass`,
	495:  `BUILD_DEVICE`,
	17:   `AppVersion`,
	689:  `year_class_cached_value_pref`,
	21:   `Fixed_21`,
	23:   `NetworkType`,
	2617: `year_class_cached_value_2016_pref`,
	1657: `Fixed_1657`,
	4473: `ab_props:sys:config_key`,
	1659: `Fixed_1659`,
	2141: `server_props:config_key`,
	287:  `BUILD_MANUFACTURER`,
}}

var WamMessageReceive = &ClassAttr{450, "WamMessageReceive", 20, map[int32]string{
	9: "isViewOnce",
	4: "messageIsInternational",
	5: "messageIsOffline",
	2: "messageMediaType",
	6: "messageReceiveT0",
	7: "messageReceiveT1",
	1: "messageType", // 1:personal 2:group 3:broadcast 4:sns
}}

// when send Ptt
var WamPtt = &ClassAttr{458, "WamPtt", 1, map[int32]string{
	5: "pttDuration",
	4: "pttLock",
	1: "pttResult",
	3: "pttSize",
	2: "pttSource",
}}

// when login
var WamLogin = &ClassAttr{460, "WamLogin", 20, map[int32]string{
	10: "androidKeystoreState",
	6:  "connectionOrigin",
	5:  "connectionT",
	1:  "loginResult",
	3:  "loginT",
	4:  "longConnect",
	8:  "passive",
	2:  "retryCount",
	7:  "sequenceStep",
	9:  "null",
}}
var WamCall = &ClassAttr{462, "WamCall", 20, map[int32]string{
	// too many...
}}
var WamUiAction = &ClassAttr{472, "WamUiAction", 10000, map[int32]string{
	//4, null);
	//2, null);
	3: "uiActionT",
	1: "uiActionType",
}}
var WamE2eMessageSend = &ClassAttr{476, "WamE2eMessageSend", 20, map[int32]string{
	5: "e2eCiphertextType",
	6: "e2eCiphertextVersion",
	4: "e2eDestination", // 0:personal 1:group 2:broadcast 3:status@broadcast(sns)
	2: "e2eFailureReason",
	8: "e2eReceiverType", // TODO
	1: "e2eSuccessful",
	9: "encRetryCount",
	7: "messageMediaType",
	3: "retryCount",
}}
var WamE2eMessageRecv = &ClassAttr{478, "WamE2eMessageRecv", 20, map[int32]string{
	5: "e2eCiphertextType",
	6: "e2eCiphertextVersion",
	4: "e2eDestination", // 0:personal 1:group 2:broadcast 3:status@broadcast(sns)
	2: "e2eFailureReason",
	8: "e2eSenderType", // always 2 (personal/group)
	1: "e2eSuccessful",
	7: "messageMediaType",
	9: "offline",
	3: "retryCount",
}}
var WamCrashLog = &ClassAttr{494, "WamCrashLog", 1, map[int32]string{
	8: "androidAppStateMetadata",
	9: "androidCrashedBuildVersion",
	3: "crashContext",
	5: "crashCount",
	2: "crashReason",
	6: "crashType",
}}
var WamGroupCreate = &ClassAttr{594, "WamGroupCreate", 20, map[int32]string{
	1: "groupCreateEntryPoint",
}}
var WamProfilePicDownload = &ClassAttr{848, "WamProfilePicDownload", 20, map[int32]string{
	1: "profilePicDownloadResult",
	4: "profilePicDownloadSize",
	3: "profilePicDownloadT",
	2: "profilePicType",
}}
var WamMessageSend = &ClassAttr{854, "WamMessageSend", 20, map[int32]string{
	25: "deviceSizeBucket",
	30: "disappearingChatInitiator",
	23: "e2eBackfill",
	21: "ephemeralityDuration",
	22: "isViewOnce",
	8:  "mediaCaptionPresent",
	4:  "messageIsForward",
	7:  "messageIsInternational",
	24: "messageIsRevoke",
	3:  "messageMediaType",
	1:  "messageSendResult",
	17: "messageSendResultIsTerminal",
	11: "messageSendT", // msg api transfer time(resp.Time - req.Time)
	2:  "messageType",  // 1:personal 2:group 3:broadcast 4:sns
	16: "resendCount",
	18: "stickerIsFirstParty",
	20: "thumbSize",
}}
var WamChatDatabaseBackupEvent = &ClassAttr{976, "WamChatDatabaseBackupEvent", 1, map[int32]string{
	//8, null
	4: "compressionRatio",
	1: "databaseBackupOverallResult",
	2: "databaseBackupVersion",
	//11, null
	6:  "freeDiskSpace",
	10: "genericBackupFailureReason",
	//7, null
	3: "msgstoreBackupSize",
	9: "sqliteVersion",
	5: "totalBackupT",
}}
var WamContactSyncEvent = &ClassAttr{1006, "WamContactSyncEvent", 100, map[int32]string{
	20: "contactSyncBusinessResponseNew",
	10: "contactSyncChangedVersionRowCount",
	19: "contactSyncDeviceResponseNew",
	14: "contactSyncErrorCode",
	16: "contactSyncFailureProtocol",
	17: "contactSyncLatency",
	12: "contactSyncNoop",
	21: "contactSyncPayResponseNew",
	6:  "contactSyncRequestClearWaSyncData",
	5:  "contactSyncRequestIsUrgent",
	15: "contactSyncRequestProtocol",
	7:  "contactSyncRequestRetryCount",
	8:  "contactSyncRequestShouldRetry",
	11: "contactSyncRequestedCount",
	13: "contactSyncResponseCount",
	18: "contactSyncStatusResponseNew",
	9:  "contactSyncSuccess",
	1:  "contactSyncType",
	//4: null);
	//3: null);
	//2: null);
}}
var WamVideoPlay = &ClassAttr{1012, "WamVideoPlay", 1000000, map[int32]string{
	4: "videoAge",
	1: "videoDuration",
	6: "videoInitialBufferingT",
	9: "videoPlayOrigin",
	//7: null);
	8: "videoPlaySurface",
	3: "videoPlayT",
	5: "videoPlayType",
	2: "videoSize",
}}
var WamAppLaunch = &ClassAttr{1094, "WamAppLaunch", 1000, map[int32]string{
	2: "appLaunchCpuT",
	7: "appLaunchDestination",
	//3: null);
	//4: null);
	1: "appLaunchT",
	5: "appLaunchTypeT",
}}
var WamAndroidMediaTranscodeEvent = &ClassAttr{1138, "WamAndroidMediaTranscodeEvent", 200, map[int32]string{
	//9:  null);
	10: "dstDurationSec",
	8:  "dstHeight",
	11: "dstSize",
	7:  "dstWidth",
	17: "durationMs",
	14: "errorType",
	1:  "fileIsDoodle",
	20: "firstScanSize",
	26: "hasStatusMessage",
	15: "isSuccess",
	24: "lowQualitySize",
	23: "maxEdge",
	//27: null);
	25: "midQualitySize",
	13: "operation",
	22: "photoCompressionQuality",
	19: "progressiveJpeg",
	4:  "srcBitrate",
	5:  "srcDurationSec",
	3:  "srcHeight",
	6:  "srcSize",
	2:  "srcWidth",
	21: "thumbnailSize",
	18: "totalQueueMs",
	16: "transcodeMediaType",
	12: "transcoderSupported",
}}
var WamDaily = &ClassAttr{1158, "WamDaily", -1, map[int32]string{
	11:  "addressbookSize",
	12:  "addressbookWhatsappSize",
	135: "androidAdvertisingId",
	37:  "androidApiLevel",
	39:  "androidHasSdCard",
	42:  "androidIsJidGoogleDriveCapable",
	41:  "androidIsJidGoogleDriveEligible",
	40:  "androidIsSdCardRemovable",
	139: "androidKeystoreState",
	98:  "androidRamLow",
	49:  "androidVideoTranscodeSupported",
	103: "appCodeHash",
	121: "appStandbyBucket",
	48:  "appUsingForcedLocale",
	90:  "autoDlAudioCellular",
	91:  "autoDlAudioRoaming",
	89:  "autoDlAudioWifi",
	96:  "autoDlDocCellular",
	97:  "autoDlDocRoaming",
	95:  "autoDlDocWifi",
	87:  "autoDlImageCellular",
	88:  "autoDlImageRoaming",
	86:  "autoDlImageWifi",
	93:  "autoDlVideoCellular",
	94:  "autoDlVideoRoaming",
	92:  "autoDlVideoWifi",
	126: "backgroundRestricted",
	10:  "backupNetworkSetting",
	138: "backupRestoreEncryptionVersion",
	9:   "backupSchedule",
	128: "bgDataRestriction",
	18:  "broadcastArchivedChatCount",
	17:  "broadcastChatCount",
	19:  "chatDatabaseSize",
	85:  "cpuAbi",
	140: "defaultDisappearingDuration",
	153: "deviceLanguage",
	109: "externalStorageAvailSize",
	110: "externalStorageTotalSize",
	112: "favoritedFirstPartyStickerCount",
	111: "favoritedTotalStickerCount",
	119: "fingerprintLockEnabled",
	62:  "googleAccountCount",
	43:  "googlePlayServicesAvailable",
	79:  "googlePlayServicesVersion",
	16:  "groupArchivedChatCount",
	15:  "groupChatCount",
	14:  "individualArchivedChatCount",
	13:  "individualChatCount",
	120: "installSource",
	137: "installedAnimatedThirdPartyStickerPackCount",
	115: "installedFirstPartyStickerPackCount",
	114: "installedTotalStickerPackCount",
	45:  "isBluestacks",
	46:  "isGenymotion",
	78:  "isMonkeyrunnerRunning",
	60:  "isRooted",
	61:  "isUsingCustomRom",
	38:  "isWhatsappPlusUser",
	154: "keyboardLanguage",
	82:  "labelsTableLabelCount",
	84:  "labelsTableLabeledContactsCount",
	83:  "labelsTableLabeledMessagesCount",
	5:   "languageCode",
	63:  "lastBackupTimestamp",
	44:  "libcQemuPresent",
	81:  "liveLocationReportingT",
	80:  "liveLocationSharingT",
	6:   "locationCode",
	21:  "mediaFolderFileCount", // /sdcard/Media
	20:  "mediaFolderSize",      // /sdcard/Media
	155: "modifiedInternalProps",
	7:   "networkIsRoaming",
	4:   "osBuildNumber",
	118: "osNotificationSetting",
	102: "packageName",
	100: "paymentsIsEnabled",
	57:  "permissionAccessCoarseLocation",
	58:  "permissionAccessFineLocation",
	56:  "permissionCamera",
	52:  "permissionGetAccounts",
	50:  "permissionReadContacts",
	53:  "permissionReadExternalStorage",
	59:  "permissionReceiveSms",
	55:  "permissionRecordAudio",
	51:  "permissionWriteContacts",
	54:  "permissionWriteExternalStorage",
	156: "phoneCores",
	8:   "receiptsEnabled",
	77:  "signatureHash",
	31:  "storageAvailSize",
	32:  "storageTotalSize",
	127: "timeDeltaSinceLastEvent",
	23:  "videoFolderFileCount",
	22:  "videoFolderSize",
}}

// first time login
var WamRegistrationComplete = &ClassAttr{1342, "WamRegistrationComplete", -1, map[int32]string{
	9:  "deviceIdentifier",
	4:  "registrationAttemptSkipWithNoVertical",
	7:  "registrationContactsPermissionSource",
	10: "registrationGoogleDriveBackupStatus",
	5:  "registrationProfilePictureSet",
	6:  "registrationProfilePictureTapped",
	3:  "registrationRetryFetchingBizProfile",
	8:  "registrationStoragePermissionSource",
	1:  "registrationT",
	2:  "registrationTForFillBusinessInfoScreen",
}}
var WamAndroidEmojiDictionaryFetch = &ClassAttr{1368, "WamAndroidEmojiDictionaryFetch", 5, map[int32]string{
	//5: null);
	4: "currentLanguages",
	6: "doNetworkFetch",
	2: "isFirstAttempt",
	1: "requestedLanguages",
	9: "result",
	7: "resultHttpCode",
	8: "resultLanguages",
	3: "timeSinceLastRequestMsT",
}}

// biz version, first register
var WamEditBusinessProfile = &ClassAttr{1466, "WamEditBusinessProfile", 1, map[int32]string{
	10: "businessProfileEntryPoint",
	2:  "editBusinessProfileSessionId",
	1:  "editProfileAction",
}}
var WamUserActivitySessionSummary = &ClassAttr{1502, "WamUserActivitySessionSummary", 20, map[int32]string{
	2: "userActivityDuration",
	5: "userActivityForeground",
	3: "userActivitySessionsLength",
	1: "userActivityStartTime",
	4: "userActivityTimeChange",
	6: "userSessionSummarySequence",
}}
var WamBannerEvent = &ClassAttr{1578, "WamBannerEvent", 20, map[int32]string{
	2: "bannerOperation",
	1: "bannerType",
}}
var WamMediaUpload2 = &ClassAttr{1588, "WamMediaUpload2", 20, map[int32]string{
	43: "connectionType",
	34: "debugMediaException",
	32: "debugMediaIp",
	33: "debugUrl",
	45: "estimatedBandwidth",
	28: "finalizeConnectT",
	31: "finalizeHttpCode",
	30: "finalizeIsReuse",
	29: "finalizeNetworkT",
	49: "isViewOnce",
	46: "mediaId",
	42: "networkStack",
	4:  "overallAttemptCount",
	10: "overallConnBlockFetchT",
	41: "overallConnectionClass",
	37: "overallCumT",
	38: "overallCumUserVisibleT",
	5:  "overallDomain",
	36: "overallIsFinal",
	16: "overallIsForward",
	13: "overallIsManual",
	//11: null);
	40: "overallMediaKeyReuse",
	7:  "overallMediaSize",
	1:  "overallMediaType",
	6:  "overallMmsVersion",
	12: "overallOptimisticFlag",
	9:  "overallQueueT",
	3:  "overallRetryCount",
	8:  "overallT",
	15: "overallTranscodeT",
	39: "overallUploadMode",
	44: "overallUploadOrigin",
	35: "overallUploadResult",
	14: "overallUserVisibleT",
	17: "resumeConnectT",
	20: "resumeHttpCode",
	19: "resumeIsReuse",
	18: "resumeNetworkT",
	27: "uploadBytesTransferred",
	22: "uploadConnectT",
	25: "uploadHttpCode",
	24: "uploadIsReuse",
	26: "uploadIsStreaming",
	23: "uploadNetworkT",
	21: "uploadResumePoint",
	//48: null);
	//47: null);

}}
var WamMediaDownload2 = &ClassAttr{1590, "WamMediaDownload2", 5, map[int32]string{
	31: "connectionType",
	24: "debugMediaException",
	22: "debugMediaIp",
	23: "debugUrl",
	20: "downloadBytesTransferred",
	15: "downloadConnectT",
	18: "downloadHttpCode",
	17: "downloadIsReuse",
	19: "downloadIsStreaming",
	16: "downloadNetworkT",
	37: "downloadQuality",
	14: "downloadResumePoint",
	21: "downloadTimeToFirstByteT",
	36: "estimatedBandwidth",
	41: "isViewOnce",
	38: "mediaId",
	30: "networkStack",
	4:  "overallAttemptCount",
	39: "overallBackendStore",
	10: "overallConnBlockFetchT",
	29: "overallConnectionClass",
	27: "overallCumT",
	5:  "overallDomain",
	11: "overallDownloadMode",
	35: "overallDownloadOrigin",
	25: "overallDownloadResult",
	13: "overallFileValidationT",
	28: "overallIsEncrypted",
	26: "overallIsFinal",
	7:  "overallMediaSize",
	1:  "overallMediaType",
	6:  "overallMmsVersion",
	9:  "overallQueueT",
	3:  "overallRetryCount",
	8:  "overallT",
	40: "usedFallbackHint",
}}

// biz version, once
var WamSmbVnameCertHealth = &ClassAttr{1602, "WamSmbVnameCertHealth", 1, map[int32]string{
	1: "smbVnameCertHealthResult",
}}
var WamChatMessageCounts = &ClassAttr{1644, "WamChatMessageCounts", 1, map[int32]string{
	15: "chatEphemeralityDuration",
	8:  "chatMuted",
	2:  "chatTypeInd",
	14: "ephemeralMessagesReceived",
	13: "ephemeralMessagesSent",
	19: "groupSize",
	6:  "isAContact",
	5:  "isAGroup",
	10: "isArchived",
	9:  "isPinned",
	4:  "messagesReceived",
	3:  "messagesSent",
	12: "messagesStarred",
	11: "messagesUnread",
	7:  "startTime",
	18: "viewOnceMessagesOpened",
	17: "viewOnceMessagesReceived",
	16: "viewOnceMessagesSent",
}}
var WamStatusDaily = &ClassAttr{1676, "WamStatusDaily", 20, map[int32]string{
	3: "statusAvailableCountDaily",
	1: "statusAvailableRowsCountDaily",
	4: "statusViewedCountDaily",
	2: "statusViewedRowsCountDaily",
}}
var WamCriticalEvent = &ClassAttr{1684, "WamCriticalEvent", 1, map[int32]string{
	2: "context",
	3: "debug",
	1: "name",
}}
var WamCatalogBiz = &ClassAttr{1722, "WamCatalogBiz", 1, map[int32]string{
	13: "cartToggle",
	1:  "catalogBizAction",
	7:  "catalogEntryPoint",
	3:  "catalogSessionId",
	8:  "deepLinkOpenFrom",
	5:  "errorCode",
	10: "isOrderMsgAttached",
	9:  "orderId",
	6:  "productCount",
	2:  "productId",
	12: "productIds",
	11: "quantity",
}}
var WamMediaDailyDataUsage = &ClassAttr{1766, "WamMediaDailyDataUsage", 1, map[int32]string{
	2:  "bytesReceived",
	1:  "bytesSent",
	13: "countDownloaded",
	14: "countForward",
	11: "countMessageReceived",
	10: "countMessageSent",
	//18, ";
	15: "countShared",
	12: "countUploaded",
	16: "countViewed",
	7:  "isAutoDownload",
	6:  "mediaTransferOrigin",
	4:  "mediaType",
	3:  "transferDate",
	5:  "transferRadio",
}}
var WamStickerPackDownload = &ClassAttr{1844, "WamStickerPackDownload", 1, map[int32]string{
	1: "stickerPackDownloadOrigin",
	2: "stickerPackIsFirstParty",
}}
var WamStickerPickerOpened = &ClassAttr{1854, "WamStickerPickerOpened", 20, map[int32]string{}}

// not verified
var WamStickerSearchOpened = &ClassAttr{1858, "WamStickerSearchOpened", 20, map[int32]string{}}

var WamAndroidDatabaseOverallMigrationEvent = &ClassAttr{1910, "WamAndroidDatabaseOverallMigrationEvent", 5, map[int32]string{
	6:  "afterMigrationMsgstoreSize",
	5:  "beforeMigrationMsgstoreSize",
	7:  "null",
	8:  "freeSpaceAvailable",
	24: "migrationInitiator",
	3:  "migrationProcessedCnt",
	2:  "migrationRegisteredCnt",
	1:  "migrationSucceeded",
	4:  "migrationT",
	23: "phaseConsistencyFailedCnt",
	22: "phaseConsistencySkippedCnt",
	21: "phaseConsistencySuccessCnt",
	14: "phaseMigrationFailedCnt",
	13: "phaseMigrationSkippedCnt",
	12: "phaseMigrationSuccessCnt",
	11: "phaseRollbackFailedCnt",
	10: "phaseRollbackSkippedCnt",
	9:  "phaseRollbackSuccessCnt",
	20: "phaseVerificationFailedCnt",
	19: "phaseVerificationSkippedCnt",
	18: "phaseVerificationSuccessCnt",
}}
var WamAndroidDatabaseMigrationEvent = &ClassAttr{1912, "WamAndroidDatabaseMigrationEvent", 20, map[int32]string{
	5:  "afterMigrationMsgstoreSize",
	4:  "beforeMigrationMsgstoreSize",
	9:  "freeSpaceAvailable",
	1:  "migrationName",
	10: "migrationSkipReason",
	2:  "migrationStatus",
	3:  "migrationT",
	6:  "retryCount",
	7:  "rowProcessedCnt",
	8:  "rowSkippedCnt",
}}
var WamAdvertisingId = &ClassAttr{1942, "WamAdvertisingId (only Samsung with GooglePlay)", 1, map[int32]string{
	1: "androidAdvertisingId",
}}

var WamAndroidMessageSendPerf = &ClassAttr{1994, "WamAndroidMessageSendPerf", 2000, map[int32]string{
	16: "appRestart",
	26: "deviceSizeBucket",
	11: "durationAbs",
	12: "durationRelative",
	1:  "durationT",
	15: "fetchPrekeys",
	21: "fetchPrekeysPercentage",
	17: "groupSizeBucket",
	9:  "isMessageFanout",
	8:  "isMessageForward",
	24: "isRevokeMessage",
	18: "jobsInQueue",
	3:  "mediaType",
	4:  "messageType",
	14: "networkWasDisconnected",
	13: "sendCount",
	10: "sendRetryCount",
	2:  "sendStage",
	23: "senderKeyDistributionCountPercentage",
	25: "sessionsMissingWhenComposing",
	20: "targetDeviceGroupSizeBucket",
	19: "threadsInExecution",
}}
var WamRegInit = &ClassAttr{2046, "WamRegInit", 1000, map[int32]string{
	2: "contactsSyncT",
	4: "groupsInitDidTimeout",
	3: "groupsInitT",
	6: "profilePhotosDownloadDidTimeout",
	5: "profilePhotosDownloadT",
	1: "totalT",
}}
var WamAndroidPerfTimer = &ClassAttr{2052, "WamAndroidPerfTimer", 20, map[int32]string{
	1: "androidPerfDuration",
	3: "androidPerfExtraData",
	2: "androidPerfName",
}}
var WamAndroidRegDirectMigrationFlow = &ClassAttr{2054, "WamAndroidRegDirectMigrationFlow", 1, map[int32]string{
	13: "didNotShowMigrationScreenWhenPossible",
	15: "didReceiveRcFromConsumer",
	17: "didSuccessfullySkipSmsVerification",
	3:  "enteredSamePhoneNumberAsSisterApp",
	4:  "firstMigrationFailureReason",
	10: "fixed_null",
	9:  "migrateMediaResult",
	8:  "migratePhoneNumberScreenAction",
	1:  "migrationDurationT",
	16: "migrationSessionId",
	2:  "migrationTotalSize",
	12: "notEnoughStorageSpaceWarningShown",
	11: "otherFilesMigrationFailed",
	14: "providerAppVersionCode",
	5:  "secondMigrationFailureReason",
	7:  "spacePredictedToNeed",
	6:  "thirdMigrationFailureReason",
}}
var WamCameraTti = &ClassAttr{2064, "WamCameraTti", 20, map[int32]string{
	4: "cameraApi",
	1: "cameraTtiDuration",
	3: "cameraType",
	2: "launchType",
}}
var WamUiActionRealTime = &ClassAttr{2098, "WamUiActionRealTime", 10000, map[int32]string{
	1: "chatdInternetConnectivity",
}}
var WamAndroidMessageTargetPerf = &ClassAttr{2170, "WamAndroidMessageTargetPerf", 20000, map[int32]string{
	1: "durationReceiptT",
	//9: null);
	3: "mediaType",
	//6: null);
	//7: null);
	//5: null);
	//4: null);
	//8: null);
	2: "targetStage",
}}
var WamAndroidHourlyCron = &ClassAttr{2198, "WamAndroidHourlyCron", 20, map[int32]string{
	2: "hourlyCronCompletedCount",
	3: "hourlyCronCountPeriod",
	1: "hourlyCronStartedCount",
}}
var WamAndroidNtpSync = &ClassAttr{2204, "WamAndroidNtpSync", 20, map[int32]string{
	4: "ntpSyncCountPeriod",
	3: "ntpSyncFailedCount",
	1: "ntpSyncStartedCount",
	2: "ntpSyncSucceededCount",
	5: "ntpSyncWorkManagerInit",
}}
var WamBusinessOnboardingInteraction = &ClassAttr{2222, "WamBusinessOnboardingInteraction", 1, map[int32]string{
	1: "businessOnboardingAction",
}}
var WamAndroidTestSchedulerApi = &ClassAttr{2232, "WamAndroidTestSchedulerApi", 20, map[int32]string{
	4:  "androidTestSchedulerAlarmApiCompleted",
	2:  "androidTestSchedulerAlarmApiScheduled",
	3:  "androidTestSchedulerAlarmApiStarted",
	7:  "androidTestSchedulerAlarmManualCompleted",
	5:  "androidTestSchedulerAlarmManualScheduled",
	6:  "androidTestSchedulerAlarmManualStarted",
	10: "androidTestSchedulerJobApiCompleted",
	8:  "androidTestSchedulerJobApiScheduled",
	9:  "androidTestSchedulerJobApiStarted",
	16: "androidTestSchedulerJobManualPostCompleted",
	14: "androidTestSchedulerJobManualPostScheduled",
	15: "androidTestSchedulerJobManualPostStarted",
	13: "androidTestSchedulerJobManualPreCompleted",
	11: "androidTestSchedulerJobManualPreScheduled",
	12: "androidTestSchedulerJobManualPreStarted",
	1:  "androidTestSchedulerPeriod",
	19: "androidTestSchedulerWorkApiCompleted",
	17: "androidTestSchedulerWorkApiScheduled",
	18: "androidTestSchedulerWorkApiStarted",
}}

// always hit, never send
var WamWamTestAnonymous0 = &ClassAttr{2240, "WamWamTestAnonymous0", 1, map[int32]string{
	// 2 attr, useless
}}
var WamSignCredential = &ClassAttr{2242, "WamSignCredential", 100, map[int32]string{
	6: "applicationState",
	4: "overallT",
	7: "projectCode",
	2: "retryCount",
	1: "signCredentialResult",
	3: "signCredentialT",
	5: "waConnectedToChatd",
}}
var WamPsBufferUpload = &ClassAttr{2244, "WamPsBufferUpload", 100, map[int32]string{
	6:  "applicationState",
	3:  "psBufferUploadHttpResponseCode",
	1:  "psBufferUploadResult",
	2:  "psBufferUploadT",
	11: "psDitheredT",
	10: "psForceUpload",
	4:  "psTokenNotReadyReason",
	9:  "psUploadReason",
	5:  "waConnectedToChatd",
}}
var WamPsIdCreate = &ClassAttr{2310, "WamPsIdCreate", 1, map[int32]string{}}
var WamChatAction = &ClassAttr{2312, "WamChatAction", 20, map[int32]string{
	3: "chatActionChatType",
	2: "chatActionEntryPoint",
	4: "chatActionMuteDuration",
	1: "chatActionType",
}}
var WamAndroidDatabaseMigrationDailyStatus = &ClassAttr{2318, "WamAndroidDatabaseMigrationDailyStatus", -1, map[int32]string{
	1:  "dbMigrationBlankMeJid",
	7:  "dbMigrationBroadcastMeJid",
	29: "dbMigrationCallLog",
	4:  "dbMigrationChat",
	36: "dbMigrationDropDeprecatedTables",
	28: "dbMigrationEphemeral",
	27: "dbMigrationEphemeralSetting",
	19: "dbMigrationFrequent",
	3:  "dbMigrationFts",
	14: "dbMigrationFuture",
	6:  "dbMigrationGroupParticipant",
	5:  "dbMigrationJid",
	10: "dbMigrationLabelJid",
	32: "dbMigrationLegacyQuotedOrderMessage",
	11: "dbMigrationLink",
	20: "dbMigrationLocation",
	25: "dbMigrationMainMessage",
	17: "dbMigrationMention",
	2:  "dbMigrationMessageMedia",
	30: "dbMigrationMessageMediaFixer",
	24: "dbMigrationMissedCall",
	22: "dbMigrationPayment",
	15: "dbMigrationQuoted",
	31: "dbMigrationQuotedOrderMessage",
	33: "dbMigrationQuotedOrderMessageV2",
	8:  "dbMigrationReceiptDevice",
	9:  "dbMigrationReceiptUser",
	35: "dbMigrationRenameDeprecatedTables",
	18: "dbMigrationRevoked",
	23: "dbMigrationSendCount",
	16: "dbMigrationSystem",
	12: "dbMigrationText",
	21: "dbMigrationThumbnail",
	13: "dbMigrationVcard",
	26: "timeSinceLastMigrationAtemptT",
}}
var WamTestAnonymousDaily = &ClassAttr{2328, "WamTestAnonymousDaily", 1, map[int32]string{}}
var WamMessageStanzaReceive = &ClassAttr{2494, "WamMessageStanzaReceive", 2000, map[int32]string{
	5:  "hasSenderKeyDistributionMessage",
	3:  "mediaType",
	10: "messageStanzaDecryptQueueSize",
	1:  "messageStanzaDuration",
	6:  "messageStanzaE2eSuccess",
	7:  "messageStanzaIsEphemeral",
	2:  "messageStanzaOfflineCount",
	8:  "messageStanzaRevoke",
	9:  "messageStanzaStage",
	4:  "messageType",
}}
var WamReceiptStanzaReceive = &ClassAttr{2496, "WamReceiptStanzaReceive", 20000, map[int32]string{
	//2: null);
	10: "messageType",
	1:  "receiptStanzaDuration",
	6:  "receiptStanzaHasOrphaned",
	3:  "receiptStanzaOfflineCount",
	8:  "receiptStanzaProcessedCount",
	5:  "receiptStanzaRetryVer",
	9:  "receiptStanzaStage",
	7:  "receiptStanzaTotalCount",
	4:  "receiptStanzaType",
}}
var WamPrekeysFetch = &ClassAttr{2540, "WamPrekeysFetch", 20, map[int32]string{
	1: "onIdentityChange",
	3: "prekeysFetchContext",
	2: "prekeysFetchCount",
}}
var WamNotificationStanzaReceive = &ClassAttr{2570, "WamNotificationStanzaReceive", 2000, map[int32]string{
	1: "notificationStanzaDuration",
	2: "notificationStanzaOfflineCount",
	4: "notificationStanzaStage",
	5: "notificationStanzaSubType",
	3: "notificationStanzaType",
}}
var WamAndroidKeystoreAuthkeySuccess = &ClassAttr{2598, "WamAndroidKeystoreAuthkeySuccess", 20, map[int32]string{
	3: "androidKeystoreState",
	2: "numFailures",
	1: "numSuccessfulReads",
}}
var WamAndroidInfraHealth = &ClassAttr{2642, "WamAndroidInfraHealth", 1, map[int32]string{
	21: "psDailyStartsBgCold",
	1:  "psDailyStartsCold",
	22: "psDailyStartsFgCold",
	3:  "psDailyStartsLukeWarm",
	2:  "psDailyStartsWarms",
	//15: null);
	//14: null);
	//16: null);
	//17: null);
	//7:  null);
	//10: null);
	//11: null);
	//5:  null);
	//8:  null);
	//12: null);
	//13: null);
	//6:  null);
	//9:  null);
	//4:  null);
	//18: null);
	19: "timeSinceLastColdStartInMin",
	20: "timeSinceLastEventInMin",
	24: "timeSinceLastLukewarmStartInMin",
	23: "timeSinceLastWarmStartInMin",
}}
var WamStickerCommonQueryToStaticServer = &ClassAttr{2740, "WamStickerCommonQueryToStaticServer", 20, map[int32]string{
	3: "androidKeystoreState",
	2: "numFailures",
	1: "numSuccessfulReads",
}}
var WamStickerDbMigrationStart = &ClassAttr{2790, "WamStickerDbMigrationStart", 20, map[int32]string{
	1: "preStickerDbMigrationRetryCount",
}}
var WamStickerDbMigrationEnd = &ClassAttr{2792, "WamStickerDbMigrationEnd", 20, map[int32]string{
	//2: null);
	1: "migrationResult",
	4: "numberOfStickersMigratedIn10s",
	5: "stickerDbMigrationRetryCount",
	3: "timeTookForMigrationInMs",
}}
var WamPttDaily = &ClassAttr{2938, "WamPttDaily", -1, map[int32]string{
	9:  "pttCancelBroadcast",
	8:  "pttCancelGroup",
	7:  "pttCancelIndividual",
	15: "pttDraftReviewBroadcast",
	14: "pttDraftReviewGroup",
	13: "pttDraftReviewIndividual",
	21: "pttFastplaybackBroadcast",
	20: "pttFastplaybackGroup",
	19: "pttFastplaybackIndividual",
	12: "pttLockBroadcast",
	11: "pttLockGroup",
	10: "pttLockIndividual",
	18: "pttPlaybackBroadcast",
	17: "pttPlaybackGroup",
	16: "pttPlaybackIndividual",
	3:  "pttRecordBroadcast",
	2:  "pttRecordGroup",
	1:  "pttRecordIndividual",
	6:  "pttSendBroadcast",
	5:  "pttSendGroup",
	4:  "pttSendIndividual",
	25: "pttStopTapBroadcast",
	26: "pttStopTapGroup",
	27: "pttStopTapIndividual",
}}

func init() {
	MapStats = map[int32]*ClassAttr{
		Wild.Id:                                    Wild,
		WamMessageReceive.Id:                       WamMessageReceive,
		WamPtt.Id:                                  WamPtt,
		WamLogin.Id:                                WamLogin,
		WamCall.Id:                                 WamCall,
		WamUiAction.Id:                             WamUiAction,
		WamE2eMessageSend.Id:                       WamE2eMessageSend,
		WamE2eMessageRecv.Id:                       WamE2eMessageRecv,
		WamCrashLog.Id:                             WamCrashLog,
		WamGroupCreate.Id:                          WamGroupCreate,
		WamProfilePicDownload.Id:                   WamProfilePicDownload,
		WamMessageSend.Id:                          WamMessageSend,
		WamChatDatabaseBackupEvent.Id:              WamChatDatabaseBackupEvent,
		WamContactSyncEvent.Id:                     WamContactSyncEvent,
		WamVideoPlay.Id:                            WamVideoPlay,
		WamAppLaunch.Id:                            WamAppLaunch,
		WamAndroidMediaTranscodeEvent.Id:           WamAndroidMediaTranscodeEvent,
		WamDaily.Id:                                WamDaily,
		WamRegistrationComplete.Id:                 WamRegistrationComplete,
		WamAndroidEmojiDictionaryFetch.Id:          WamAndroidEmojiDictionaryFetch,
		WamEditBusinessProfile.Id:                  WamEditBusinessProfile,
		WamUserActivitySessionSummary.Id:           WamUserActivitySessionSummary,
		WamBannerEvent.Id:                          WamBannerEvent,
		WamMediaUpload2.Id:                         WamMediaUpload2,
		WamMediaDownload2.Id:                       WamMediaDownload2,
		WamSmbVnameCertHealth.Id:                   WamSmbVnameCertHealth,
		WamChatMessageCounts.Id:                    WamChatMessageCounts,
		WamStatusDaily.Id:                          WamStatusDaily,
		WamCriticalEvent.Id:                        WamCriticalEvent,
		WamCatalogBiz.Id:                           WamCatalogBiz,
		WamMediaDailyDataUsage.Id:                  WamMediaDailyDataUsage,
		WamStickerPackDownload.Id:                  WamStickerPackDownload,
		WamStickerPickerOpened.Id:                  WamStickerPickerOpened,
		WamStickerSearchOpened.Id:                  WamStickerSearchOpened,
		WamAndroidDatabaseOverallMigrationEvent.Id: WamAndroidDatabaseOverallMigrationEvent,
		WamAndroidDatabaseMigrationEvent.Id:        WamAndroidDatabaseMigrationEvent,
		WamAdvertisingId.Id:                        WamAdvertisingId,
		WamAndroidMessageSendPerf.Id:               WamAndroidMessageSendPerf,
		WamRegInit.Id:                              WamRegInit,
		WamAndroidPerfTimer.Id:                     WamAndroidPerfTimer,
		WamAndroidRegDirectMigrationFlow.Id:        WamAndroidRegDirectMigrationFlow,
		WamCameraTti.Id:                            WamCameraTti,
		WamUiActionRealTime.Id:                     WamUiActionRealTime,
		WamAndroidMessageTargetPerf.Id:             WamAndroidMessageTargetPerf,
		WamAndroidHourlyCron.Id:                    WamAndroidHourlyCron,
		WamAndroidNtpSync.Id:                       WamAndroidNtpSync,
		WamBusinessOnboardingInteraction.Id:        WamBusinessOnboardingInteraction,
		WamAndroidTestSchedulerApi.Id:              WamAndroidTestSchedulerApi,
		WamWamTestAnonymous0.Id:                    WamWamTestAnonymous0,
		WamSignCredential.Id:                       WamSignCredential,
		WamPsBufferUpload.Id:                       WamPsBufferUpload,
		WamPsIdCreate.Id:                           WamPsIdCreate,
		WamChatAction.Id:                           WamChatAction,
		WamAndroidDatabaseMigrationDailyStatus.Id:  WamAndroidDatabaseMigrationDailyStatus,
		WamTestAnonymousDaily.Id:                   WamTestAnonymousDaily,
		WamMessageStanzaReceive.Id:                 WamMessageStanzaReceive,
		WamReceiptStanzaReceive.Id:                 WamReceiptStanzaReceive,
		WamPrekeysFetch.Id:                         WamPrekeysFetch,
		WamNotificationStanzaReceive.Id:            WamNotificationStanzaReceive,
		WamAndroidKeystoreAuthkeySuccess.Id:        WamAndroidKeystoreAuthkeySuccess,
		WamAndroidInfraHealth.Id:                   WamAndroidInfraHealth,
		WamStickerCommonQueryToStaticServer.Id:     WamStickerCommonQueryToStaticServer,
		WamStickerDbMigrationStart.Id:              WamStickerDbMigrationStart,
		WamStickerDbMigrationEnd.Id:                WamStickerDbMigrationEnd,
		WamPttDaily.Id:                             WamPttDaily,
	}

}
