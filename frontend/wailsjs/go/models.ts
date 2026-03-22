export namespace app {
	
	export class CacheInfo {
	    download_count: number;
	    completed_count: number;
	    pending_count: number;
	    video_count: number;
	    total_library_size: number;
	    orphaned_files_count: number;
	    orphaned_files_size: number;
	
	    static createFrom(source: any = {}) {
	        return new CacheInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.download_count = source["download_count"];
	        this.completed_count = source["completed_count"];
	        this.pending_count = source["pending_count"];
	        this.video_count = source["video_count"];
	        this.total_library_size = source["total_library_size"];
	        this.orphaned_files_count = source["orphaned_files_count"];
	        this.orphaned_files_size = source["orphaned_files_size"];
	    }
	}
	export class DownloadResult {
	    id: string;
	    url: string;
	    status: string;
	    progress: number;
	    title: string;
	    channel: string;
	    thumbnail_url: string;
	    format_id: string;
	    quality: string;
	    error_message: string;
	    youtube_id: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.title = source["title"];
	        this.channel = source["channel"];
	        this.thumbnail_url = source["thumbnail_url"];
	        this.format_id = source["format_id"];
	        this.quality = source["quality"];
	        this.error_message = source["error_message"];
	        this.youtube_id = source["youtube_id"];
	    }
	}
	export class EditJobResult {
	    id: string;
	    source_video_id: string;
	    output_video_id?: string;
	    status: string;
	    operation: string;
	    progress: number;
	    error_message?: string;
	    created_at: number;
	    completed_at?: number;
	
	    static createFrom(source: any = {}) {
	        return new EditJobResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.source_video_id = source["source_video_id"];
	        this.output_video_id = source["output_video_id"];
	        this.status = source["status"];
	        this.operation = source["operation"];
	        this.progress = source["progress"];
	        this.error_message = source["error_message"];
	        this.created_at = source["created_at"];
	        this.completed_at = source["completed_at"];
	    }
	}
	export class EditSettingsInput {
	    crop_start?: number;
	    crop_end?: number;
	    crop_x?: number;
	    crop_y?: number;
	    crop_width?: number;
	    crop_height?: number;
	    watermark_type?: string;
	    watermark_text?: string;
	    watermark_image?: string;
	    watermark_position?: string;
	    watermark_opacity?: number;
	    watermark_size?: number;
	    output_format?: string;
	    output_codec?: string;
	    output_quality?: number;
	    output_resolution?: string;
	    brightness?: number;
	    contrast?: number;
	    saturation?: number;
	    rotation?: number;
	    speed?: number;
	    volume?: number;
	    remove_audio?: boolean;
	    output_filename?: string;
	    replace_original?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EditSettingsInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.crop_start = source["crop_start"];
	        this.crop_end = source["crop_end"];
	        this.crop_x = source["crop_x"];
	        this.crop_y = source["crop_y"];
	        this.crop_width = source["crop_width"];
	        this.crop_height = source["crop_height"];
	        this.watermark_type = source["watermark_type"];
	        this.watermark_text = source["watermark_text"];
	        this.watermark_image = source["watermark_image"];
	        this.watermark_position = source["watermark_position"];
	        this.watermark_opacity = source["watermark_opacity"];
	        this.watermark_size = source["watermark_size"];
	        this.output_format = source["output_format"];
	        this.output_codec = source["output_codec"];
	        this.output_quality = source["output_quality"];
	        this.output_resolution = source["output_resolution"];
	        this.brightness = source["brightness"];
	        this.contrast = source["contrast"];
	        this.saturation = source["saturation"];
	        this.rotation = source["rotation"];
	        this.speed = source["speed"];
	        this.volume = source["volume"];
	        this.remove_audio = source["remove_audio"];
	        this.output_filename = source["output_filename"];
	        this.replace_original = source["replace_original"];
	    }
	}
	export class EditPresetResult {
	    id: string;
	    name: string;
	    description: string;
	    icon: string;
	    operation: string;
	    settings: EditSettingsInput;
	
	    static createFrom(source: any = {}) {
	        return new EditPresetResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	        this.operation = source["operation"];
	        this.settings = this.convertValues(source["settings"], EditSettingsInput);
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
	
	export class FFmpegCheckResult {
	    installed: boolean;
	    version: string;
	    path: string;
	    canAutoInstall: boolean;
	    installMethod: string;
	    installCommand: string;
	    installGuide: string;
	    downloadURL: string;
	    requiresAdmin: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FFmpegCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.version = source["version"];
	        this.path = source["path"];
	        this.canAutoInstall = source["canAutoInstall"];
	        this.installMethod = source["installMethod"];
	        this.installCommand = source["installCommand"];
	        this.installGuide = source["installGuide"];
	        this.downloadURL = source["downloadURL"];
	        this.requiresAdmin = source["requiresAdmin"];
	    }
	}
	export class InstallGuide {
	    title: string;
	    description: string;
	    steps: string[];
	    command: string;
	    commandDescription: string;
	    alternativeURL: string;
	    tips: string[];
	
	    static createFrom(source: any = {}) {
	        return new InstallGuide(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.description = source["description"];
	        this.steps = source["steps"];
	        this.command = source["command"];
	        this.commandDescription = source["commandDescription"];
	        this.alternativeURL = source["alternativeURL"];
	        this.tips = source["tips"];
	    }
	}
	export class ListVideosOptions {
	    search: string;
	    channel: string;
	    sort_by: string;
	    sort_desc: boolean;
	    limit: number;
	    offset: number;
	
	    static createFrom(source: any = {}) {
	        return new ListVideosOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.search = source["search"];
	        this.channel = source["channel"];
	        this.sort_by = source["sort_by"];
	        this.sort_desc = source["sort_desc"];
	        this.limit = source["limit"];
	        this.offset = source["offset"];
	    }
	}
	export class LogEntry {
	    timestamp: string;
	    level: string;
	    component: string;
	    message: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.level = source["level"];
	        this.component = source["component"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class VideoInfoResult {
	    id: string;
	    title: string;
	    channel: string;
	    channel_id: string;
	    duration: number;
	    description: string;
	    thumbnail: string;
	    formats: ytdl.FormatInfo[];
	
	    static createFrom(source: any = {}) {
	        return new VideoInfoResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.channel = source["channel"];
	        this.channel_id = source["channel_id"];
	        this.duration = source["duration"];
	        this.description = source["description"];
	        this.thumbnail = source["thumbnail"];
	        this.formats = this.convertValues(source["formats"], ytdl.FormatInfo);
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
	export class VideoMetadataResult {
	    duration: number;
	    width: number;
	    height: number;
	    fps: number;
	    bitrate: number;
	    codec: string;
	    audio_codec?: string;
	    has_audio: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VideoMetadataResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.duration = source["duration"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.fps = source["fps"];
	        this.bitrate = source["bitrate"];
	        this.codec = source["codec"];
	        this.audio_codec = source["audio_codec"];
	        this.has_audio = source["has_audio"];
	    }
	}
	export class VideoResult {
	    id: string;
	    youtube_id: string;
	    title: string;
	    channel: string;
	    channel_id: string;
	    duration: number;
	    description: string;
	    thumbnail_url: string;
	    file_path: string;
	    file_size: number;
	    format: string;
	    quality: string;
	    downloaded_at: number;
	    watch_position: number;
	    watch_count: number;
	
	    static createFrom(source: any = {}) {
	        return new VideoResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.youtube_id = source["youtube_id"];
	        this.title = source["title"];
	        this.channel = source["channel"];
	        this.channel_id = source["channel_id"];
	        this.duration = source["duration"];
	        this.description = source["description"];
	        this.thumbnail_url = source["thumbnail_url"];
	        this.file_path = source["file_path"];
	        this.file_size = source["file_size"];
	        this.format = source["format"];
	        this.quality = source["quality"];
	        this.downloaded_at = source["downloaded_at"];
	        this.watch_position = source["watch_position"];
	        this.watch_count = source["watch_count"];
	    }
	}

}

export namespace config {
	
	export class DownloadPreset {
	    id: string;
	    name: string;
	    format: string;
	    quality: string;
	    extension: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadPreset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.format = source["format"];
	        this.quality = source["quality"];
	        this.extension = source["extension"];
	    }
	}
	export class Config {
	    user_selected_path: string;
	    download_path: string;
	    max_concurrent_downloads: number;
	    default_quality: string;
	    filename_template: string;
	    theme: string;
	    accent_color: string;
	    sidebar_collapsed: boolean;
	    default_volume: number;
	    remember_position: boolean;
	    speed_limit_kbps?: number;
	    proxy_url?: string;
	    log_path: string;
	    log_export_path: string;
	    max_log_sessions: number;
	    download_presets: DownloadPreset[];
	    ffmpeg_path?: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.user_selected_path = source["user_selected_path"];
	        this.download_path = source["download_path"];
	        this.max_concurrent_downloads = source["max_concurrent_downloads"];
	        this.default_quality = source["default_quality"];
	        this.filename_template = source["filename_template"];
	        this.theme = source["theme"];
	        this.accent_color = source["accent_color"];
	        this.sidebar_collapsed = source["sidebar_collapsed"];
	        this.default_volume = source["default_volume"];
	        this.remember_position = source["remember_position"];
	        this.speed_limit_kbps = source["speed_limit_kbps"];
	        this.proxy_url = source["proxy_url"];
	        this.log_path = source["log_path"];
	        this.log_export_path = source["log_export_path"];
	        this.max_log_sessions = source["max_log_sessions"];
	        this.download_presets = this.convertValues(source["download_presets"], DownloadPreset);
	        this.ffmpeg_path = source["ffmpeg_path"];
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

export namespace db {
	
	export class Download {
	    id: string;
	    url: string;
	    status: string;
	    progress: number;
	    title?: string;
	    channel?: string;
	    thumbnail_url?: string;
	    format_id?: string;
	    quality?: string;
	    duration?: number;
	    error_message?: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    started_at?: any;
	    // Go type: time
	    completed_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new Download(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.title = source["title"];
	        this.channel = source["channel"];
	        this.thumbnail_url = source["thumbnail_url"];
	        this.format_id = source["format_id"];
	        this.quality = source["quality"];
	        this.duration = source["duration"];
	        this.error_message = source["error_message"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.completed_at = this.convertValues(source["completed_at"], null);
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

export namespace ytdl {
	
	export class FormatInfo {
	    format_id: string;
	    ext: string;
	    resolution: string;
	    fps: number;
	    vcodec: string;
	    acodec: string;
	    filesize: number;
	    quality: string;
	
	    static createFrom(source: any = {}) {
	        return new FormatInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format_id = source["format_id"];
	        this.ext = source["ext"];
	        this.resolution = source["resolution"];
	        this.fps = source["fps"];
	        this.vcodec = source["vcodec"];
	        this.acodec = source["acodec"];
	        this.filesize = source["filesize"];
	        this.quality = source["quality"];
	    }
	}

}

