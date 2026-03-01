import { env } from '$env/dynamic/public';
import { toast } from 'svelte-sonner';
import { goto } from '$app/navigation';
import type {
	AllAdvancementsPayload,
	BlindPayload,
	BoatPayload,
	DivinePayload,
	InformationMessagesPayload,
	PresencePayload,
	RoomClosedPayload,
	RoomMessage,
	StrongholdPayload,
} from '../types/stronghold';

export class RoomConnection {
	private code: string = '';
	private isEmbed: boolean = false;
	ws: WebSocket | null = null;
	status = $state<'connecting' | 'connected' | 'reconnecting' | 'offline' | 'error'>('offline');
	runnerStatus = $state<'online' | 'offline' | 'unknown'>('unknown');
	roomClosed = $state(false);

	// Individual Payloads
	stronghold = $state<StrongholdPayload | null>(null);
	boat = $state<BoatPayload | null>(null);
	blind = $state<BlindPayload | null>(null);
	divine = $state<DivinePayload | null>(null);
	allAdvancements = $state<AllAdvancementsPayload | null>(null);
	informationMessages = $state<InformationMessagesPayload | null>(null);

	// Ticker for connection health
	private reconnectFailures = 0;
	private maxReconnectFailures = 15;
	private explicitClose = false;

	constructor(code: string, isEmbed: boolean = false) {
		this.code = code;
		this.isEmbed = isEmbed;

		if (typeof window !== 'undefined') {
			window.addEventListener('beforeunload', () => {
				this.disconnect();
			});
		}
	}

	connect() {
		this.explicitClose = false;
		if (
			this.ws &&
			(this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)
		)
			return;

		this.status = this.reconnectFailures > 0 ? 'reconnecting' : 'connecting';

		const wsUrl = env.PUBLIC_WS_URL || 'ws://localhost:8080/ws';
		this.ws = new WebSocket(`${wsUrl}?room=${this.code}`);

		this.ws.onopen = () => {
			this.status = 'connected';
			this.reconnectFailures = 0;
		};

		this.ws.onclose = (event: CloseEvent) => {
			if (this.explicitClose) {
				this.status = 'offline';
				return;
			}

			// code 4000 = room closed by server
			if (event.code === 4000) {
				this.roomClosed = true;
				this.runnerStatus = 'offline';
				this.status = 'offline';
				return;
			}

			this.handleReconnect();
		};

		this.ws.onerror = (e) => {
			console.error('WebSocket Error', e);
		};

		this.ws.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data) as RoomMessage;

				switch (data.type) {
					case 'presence':
						if ('status' in data.payload) {
							this.runnerStatus = (data.payload as PresencePayload).status;
						}
						break;
					case 'stronghold_update':
						this.stronghold = data.payload as StrongholdPayload;
						break;
					case 'boat_update':
						this.boat = data.payload as BoatPayload;
						break;
					case 'blind_update':
						this.blind = data.payload as BlindPayload;
						break;
					case 'divine_update':
						this.divine = data.payload as DivinePayload;
						break;
					case 'all_advancements_update':
						this.allAdvancements = data.payload as AllAdvancementsPayload;
						break;
					case 'information_messages_update':
						this.informationMessages = data.payload as InformationMessagesPayload;
						break;
					case 'room_closed':
						this.runnerStatus = 'offline';
						this.roomClosed = true;
						if (!this.isEmbed) {
							const payload = data.payload as RoomClosedPayload;
							toast.error(payload.message || 'The room has been closed');
							goto('/');
						}
						break;
				}
			} catch (e) {
				console.error('Failed to parse WebSocket message', e);
			}
		};
	}

	private handleReconnect() {
		if (this.explicitClose) return;
		if (this.roomClosed) {
			this.status = 'offline';
			return;
		}

		this.ws = null;

		if (this.reconnectFailures >= this.maxReconnectFailures) {
			this.status = 'error';
			return;
		}

		this.reconnectFailures++;
		this.status = 'reconnecting';

		// Exponential backoff
		const delayMs = Math.min(1000 * Math.pow(1.5, this.reconnectFailures), 10000);
		setTimeout(() => this.connect(), delayMs);
	}

	disconnect() {
		this.explicitClose = true;
		if (this.ws) {
			this.ws.close();
			this.ws = null;
		}
		this.status = 'offline';
	}
}
