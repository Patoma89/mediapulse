export namespace config {
	
	export class Config {
	    language: string;
	    theme: string;
	    preferredEncoder: string;
	    defaultOutputDir: string;
	    customFfmpegDir: string;
	    openFolderOnDone: boolean;
	    defaultPreset: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.theme = source["theme"];
	        this.preferredEncoder = source["preferredEncoder"];
	        this.defaultOutputDir = source["defaultOutputDir"];
	        this.customFfmpegDir = source["customFfmpegDir"];
	        this.openFolderOnDone = source["openFolderOnDone"];
	        this.defaultPreset = source["defaultPreset"];
	    }
	}

}

export namespace ffmpeg {
	
	export class Binaries {
	    ffmpegPath: string;
	    ffprobePath: string;
	    source: string;
	    version: string;
	    ready: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new Binaries(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ffmpegPath = source["ffmpegPath"];
	        this.ffprobePath = source["ffprobePath"];
	        this.source = source["source"];
	        this.version = source["version"];
	        this.ready = source["ready"];
	        this.error = source["error"];
	    }
	}
	export class EncoderCapability {
	    name: string;
	    codec: string;
	    backend: string;
	    label: string;
	    available: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EncoderCapability(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.codec = source["codec"];
	        this.backend = source["backend"];
	        this.label = source["label"];
	        this.available = source["available"];
	    }
	}
	export class HardwareInfo {
	    gpuName: string;
	    encoders: EncoderCapability[];
	    recommended: string;
	
	    static createFrom(source: any = {}) {
	        return new HardwareInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gpuName = source["gpuName"];
	        this.encoders = this.convertValues(source["encoders"], EncoderCapability);
	        this.recommended = source["recommended"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MediaInfo {
	    path: string;
	    fileName: string;
	    extension: string;
	    sizeBytes: number;
	    durationSec: number;
	    bitrateKbps: number;
	    container: string;
	    hasVideo: boolean;
	    videoCodec: string;
	    videoProfile: string;
	    pixelFormat: string;
	    width: number;
	    height: number;
	    fps: number;
	    bitDepth: number;
	    hasAudio: boolean;
	    audioCodec: string;
	    audioChannels: number;
	    audioSampleRate: number;
	    audioBitrateKbps: number;
	
	    static createFrom(source: any = {}) {
	        return new MediaInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.fileName = source["fileName"];
	        this.extension = source["extension"];
	        this.sizeBytes = source["sizeBytes"];
	        this.durationSec = source["durationSec"];
	        this.bitrateKbps = source["bitrateKbps"];
	        this.container = source["container"];
	        this.hasVideo = source["hasVideo"];
	        this.videoCodec = source["videoCodec"];
	        this.videoProfile = source["videoProfile"];
	        this.pixelFormat = source["pixelFormat"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.fps = source["fps"];
	        this.bitDepth = source["bitDepth"];
	        this.hasAudio = source["hasAudio"];
	        this.audioCodec = source["audioCodec"];
	        this.audioChannels = source["audioChannels"];
	        this.audioSampleRate = source["audioSampleRate"];
	        this.audioBitrateKbps = source["audioBitrateKbps"];
	    }
	}

}

export namespace history {
	
	export class Entry {
	    id: string;
	    timestamp: string;
	    inputName: string;
	    inputPath: string;
	    outputName: string;
	    outputPath: string;
	    presetId: string;
	    encoder: string;
	    inputBytes: number;
	    outputBytes: number;
	    savedBytes: number;
	    processSec: number;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.inputName = source["inputName"];
	        this.inputPath = source["inputPath"];
	        this.outputName = source["outputName"];
	        this.outputPath = source["outputPath"];
	        this.presetId = source["presetId"];
	        this.encoder = source["encoder"];
	        this.inputBytes = source["inputBytes"];
	        this.outputBytes = source["outputBytes"];
	        this.savedBytes = source["savedBytes"];
	        this.processSec = source["processSec"];
	    }
	}
	export class Summary {
	    totalJobs: number;
	    totalInputBytes: number;
	    totalOutputBytes: number;
	    totalSavedBytes: number;
	    entries: Entry[];
	
	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalJobs = source["totalJobs"];
	        this.totalInputBytes = source["totalInputBytes"];
	        this.totalOutputBytes = source["totalOutputBytes"];
	        this.totalSavedBytes = source["totalSavedBytes"];
	        this.entries = this.convertValues(source["entries"], Entry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class AddFilesResult {
	    added: queue.Job[];
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new AddFilesResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.added = this.convertValues(source["added"], queue.Job);
	        this.errors = source["errors"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AppInfo {
	    name: string;
	    version: string;
	    os: string;
	    arch: string;
	    configPath: string;
	    zeroTelemetry: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.configPath = source["configPath"];
	        this.zeroTelemetry = source["zeroTelemetry"];
	    }
	}
	export class SystemStatus {
	    binaries: ffmpeg.Binaries;
	    hardware: ffmpeg.HardwareInfo;
	
	    static createFrom(source: any = {}) {
	        return new SystemStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.binaries = this.convertValues(source["binaries"], ffmpeg.Binaries);
	        this.hardware = this.convertValues(source["hardware"], ffmpeg.HardwareInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace presets {
	
	export class TVStrategy {
	    mode: string;
	    reason: string;
	    estimatedSpeedLabel: string;
	
	    static createFrom(source: any = {}) {
	        return new TVStrategy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.reason = source["reason"];
	        this.estimatedSpeedLabel = source["estimatedSpeedLabel"];
	    }
	}
	export class ConversionPlan {
	    outputPath: string;
	    extension: string;
	    usesTwoPass: boolean;
	    effectiveEncoder: string;
	    encoderBackend: string;
	    tvStrategy?: TVStrategy;
	    videoBitrateKbps: number;
	    audioBitrateKbps: number;
	    effectiveDurationSec: number;
	    estimatedSizeMb: number;
	    warning?: string;
	    pass1Args?: string[];
	    pass2Args?: string[];
	    singleArgs?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ConversionPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.outputPath = source["outputPath"];
	        this.extension = source["extension"];
	        this.usesTwoPass = source["usesTwoPass"];
	        this.effectiveEncoder = source["effectiveEncoder"];
	        this.encoderBackend = source["encoderBackend"];
	        this.tvStrategy = this.convertValues(source["tvStrategy"], TVStrategy);
	        this.videoBitrateKbps = source["videoBitrateKbps"];
	        this.audioBitrateKbps = source["audioBitrateKbps"];
	        this.effectiveDurationSec = source["effectiveDurationSec"];
	        this.estimatedSizeMb = source["estimatedSizeMb"];
	        this.warning = source["warning"];
	        this.pass1Args = source["pass1Args"];
	        this.pass2Args = source["pass2Args"];
	        this.singleArgs = source["singleArgs"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConvertOptions {
	    presetId: string;
	    targetSizeMb: number;
	    outputFormat: string;
	    videoCodec: string;
	    qualityCrf: number;
	    resolution: string;
	    fps: number;
	    rotation: number;
	    mute: boolean;
	    trimStartSec: number;
	    trimEndSec: number;
	    audioFormat: string;
	    audioBitrateKbps: number;
	    encoderMode: string;
	    outputDir: string;
	
	    static createFrom(source: any = {}) {
	        return new ConvertOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.presetId = source["presetId"];
	        this.targetSizeMb = source["targetSizeMb"];
	        this.outputFormat = source["outputFormat"];
	        this.videoCodec = source["videoCodec"];
	        this.qualityCrf = source["qualityCrf"];
	        this.resolution = source["resolution"];
	        this.fps = source["fps"];
	        this.rotation = source["rotation"];
	        this.mute = source["mute"];
	        this.trimStartSec = source["trimStartSec"];
	        this.trimEndSec = source["trimEndSec"];
	        this.audioFormat = source["audioFormat"];
	        this.audioBitrateKbps = source["audioBitrateKbps"];
	        this.encoderMode = source["encoderMode"];
	        this.outputDir = source["outputDir"];
	    }
	}
	export class PresetDefinition {
	    id: string;
	    nameEs: string;
	    nameEn: string;
	    descEs: string;
	    descEn: string;
	    badge: string;
	    defaultFormat: string;
	    defaultSizeMb: number;
	    accent: string;
	
	    static createFrom(source: any = {}) {
	        return new PresetDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nameEs = source["nameEs"];
	        this.nameEn = source["nameEn"];
	        this.descEs = source["descEs"];
	        this.descEn = source["descEn"];
	        this.badge = source["badge"];
	        this.defaultFormat = source["defaultFormat"];
	        this.defaultSizeMb = source["defaultSizeMb"];
	        this.accent = source["accent"];
	    }
	}

}

export namespace queue {
	
	export class Job {
	    id: string;
	    info: ffmpeg.MediaInfo;
	    options: presets.ConvertOptions;
	    plan: presets.ConversionPlan;
	    status: string;
	    progressPercent: number;
	    speed: string;
	    etaSec: number;
	    pass: number;
	    outputPath: string;
	    outputBytes: number;
	    processSec: number;
	    fallbackNotice?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.info = this.convertValues(source["info"], ffmpeg.MediaInfo);
	        this.options = this.convertValues(source["options"], presets.ConvertOptions);
	        this.plan = this.convertValues(source["plan"], presets.ConversionPlan);
	        this.status = source["status"];
	        this.progressPercent = source["progressPercent"];
	        this.speed = source["speed"];
	        this.etaSec = source["etaSec"];
	        this.pass = source["pass"];
	        this.outputPath = source["outputPath"];
	        this.outputBytes = source["outputBytes"];
	        this.processSec = source["processSec"];
	        this.fallbackNotice = source["fallbackNotice"];
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class QueueState {
	    jobs: Job[];
	    running: boolean;
	    doneCount: number;
	    pending: number;
	
	    static createFrom(source: any = {}) {
	        return new QueueState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.jobs = this.convertValues(source["jobs"], Job);
	        this.running = source["running"];
	        this.doneCount = source["doneCount"];
	        this.pending = source["pending"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace updater {
	
	export class UpdateCheckResult {
	    currentVersion: string;
	    latestVersion: string;
	    updateAvailable: boolean;
	    releaseUrl: string;
	    releaseName?: string;
	    publishedAt?: string;
	    releaseNotes?: string;
	    checkedAt: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.updateAvailable = source["updateAvailable"];
	        this.releaseUrl = source["releaseUrl"];
	        this.releaseName = source["releaseName"];
	        this.publishedAt = source["publishedAt"];
	        this.releaseNotes = source["releaseNotes"];
	        this.checkedAt = source["checkedAt"];
	        this.error = source["error"];
	    }
	}

}

