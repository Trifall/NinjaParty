import { env } from '$env/dynamic/public';
import { toast } from 'svelte-sonner';
import { SvelteMap } from 'svelte/reactivity';
import { goto } from '$app/navigation';
import type {
	AllAdvancementsPayload,
	BlindPayload,
	BoatPayload,
	DivinePayload,
	InformationMessagesPayload,
	PresencePayload,
	RoomClosedPayload,
	RoomInfoPayload,
	RoomMessage,
	StrongholdPayload,
} from '../types/stronghold';

export class RunnerState {
	name: string;
	status = $state<'online' | 'offline'>('offline');
	stronghold = $state<StrongholdPayload | null>(null);
	boat = $state<BoatPayload | null>(null);
	blind = $state<BlindPayload | null>(null);
	divine = $state<DivinePayload | null>(null);
	allAdvancements = $state<AllAdvancementsPayload | null>(null);
	informationMessages = $state<InformationMessagesPayload | null>(null);
	lastUpdate = $state(Date.now());

	constructor(name: string) {
		this.name = name;
	}
}

export class RoomConnection {
	private code: string = '';
	private isEmbed: boolean = false;
	ws: WebSocket | null = null;
	status = $state<'connecting' | 'connected' | 'reconnecting' | 'offline' | 'error'>('offline');
	roomClosed = $state(false);
	createdAt = $state<number | null>(null);

	runners = new SvelteMap<string, RunnerState>();
	runnerJoinOrder: string[] = [];

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

	get sortedRunners(): RunnerState[] {
		const runnerList: RunnerState[] = [];
		this.runners.forEach((runner) => {
			runnerList.push(runner);
		});

		return runnerList.sort((a, b) => {
			if (a.status !== b.status) {
				return a.status === 'online' ? -1 : 1;
			}
			const aIndex = this.runnerJoinOrder.indexOf(a.name);
			const bIndex = this.runnerJoinOrder.indexOf(b.name);
			if (aIndex !== bIndex) {
				return aIndex - bIndex;
			}
			return a.name.localeCompare(b.name);
		});
	}

	private getOrCreateRunner(name: string): RunnerState {
		let runner = this.runners.get(name);
		if (!runner) {
			runner = new RunnerState(name);
			this.runners.set(name, runner);
			this.runnerJoinOrder.push(name);
		}
		return runner;
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
			console.log('WebSocket connected to room:', this.code);
			this.status = 'connected';
			this.reconnectFailures = 0;
		};

		this.ws.onclose = (event: CloseEvent) => {
			if (this.explicitClose) {
				this.status = 'offline';
				return;
			}

			if (event.code === 4000) {
				this.roomClosed = true;
				this.status = 'offline';
				this.runners.forEach((runner) => {
					runner.status = 'offline';
				});
				if (!this.isEmbed && !this.roomClosed) {
					toast.error('The room has been closed');
					goto('/');
				}
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
				// handle server-level messages before runnerName check
				if (data.type === 'room_info') {
					this.createdAt = (data.payload as RoomInfoPayload).createdAt;
					return;
				}

				console.log('WS message received:', data.type);
				const runnerName = data.runnerName;

				if (!runnerName) {
					console.warn('Received message without runnerName', data);
					return;
				}

				const runner = this.getOrCreateRunner(runnerName);
				console.log(
					'Processing message for runner:',
					runnerName,
					'status:',
					runner.status,
					'runners count:',
					this.runners.size
				);

				switch (data.type) {
					case 'presence':
						if ('status' in data.payload) {
							runner.status = (data.payload as PresencePayload).status;
							runner.lastUpdate = Date.now();
						}
						break;
					case 'stronghold_update':
						runner.stronghold = data.payload as StrongholdPayload;
						runner.lastUpdate = Date.now();
						break;
					case 'boat_update':
						runner.boat = data.payload as BoatPayload;
						runner.lastUpdate = Date.now();
						break;
					case 'blind_update':
						runner.blind = data.payload as BlindPayload;
						runner.lastUpdate = Date.now();
						break;
					case 'divine_update':
						runner.divine = data.payload as DivinePayload;
						runner.lastUpdate = Date.now();
						break;
					case 'all_advancements_update':
						runner.allAdvancements = data.payload as AllAdvancementsPayload;
						runner.lastUpdate = Date.now();
						break;
					case 'information_messages_update':
						runner.informationMessages = data.payload as InformationMessagesPayload;
						runner.lastUpdate = Date.now();
						break;
					case 'room_closed':
						runner.status = 'offline';
						if (!this.roomClosed) {
							this.roomClosed = true;
							if (!this.isEmbed) {
								const payload = data.payload as RoomClosedPayload;
								toast.error(payload.message || 'The room has been closed');
								goto('/');
							}
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
