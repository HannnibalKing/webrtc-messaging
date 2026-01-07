// WebRTC utility functions
class WebRTCManager {
	constructor() {
		this.peerConnections = new Map();
		this.localStream = null;
		this.dataChannels = new Map();
		this.config = {
			iceServers: [
				{ urls: ['stun:stun.l.google.com:19302'] },
				{ urls: ['stun:stun1.l.google.com:19302'] }
			]
		};
	}

	// Get user media (audio and video)
	async getLocalStream(audio = true, video = true) {
		try {
			this.localStream = await navigator.mediaDevices.getUserMedia({
				audio: audio,
				video: video ? { width: { ideal: 1280 }, height: { ideal: 720 } } : false
			});
			return this.localStream;
		} catch (error) {
			console.error('Error accessing media devices:', error);
			throw error;
		}
	}

	// Stop local stream
	stopLocalStream() {
		if (this.localStream) {
			this.localStream.getTracks().forEach(track => track.stop());
			this.localStream = null;
		}
	}

	// Create peer connection
	createPeerConnection(peerId) {
		if (this.peerConnections.has(peerId)) {
			return this.peerConnections.get(peerId);
		}

		const peerConnection = new RTCPeerConnection(this.config);

		// Add local stream tracks
		if (this.localStream) {
			this.localStream.getTracks().forEach(track => {
				peerConnection.addTrack(track, this.localStream);
			});
		}

		// Handle ICE candidates
		peerConnection.onicecandidate = (event) => {
			if (event.candidate) {
				this.onICECandidate(peerId, event.candidate);
			}
		};

		// Handle remote stream
		peerConnection.ontrack = (event) => {
			this.onRemoteStream(peerId, event.streams[0]);
		};

		// Handle connection state changes
		peerConnection.onconnectionstatechange = () => {
			this.onConnectionStateChange(peerId, peerConnection.connectionState);
		};

		// Handle data channels
		peerConnection.ondatachannel = (event) => {
			this.setupDataChannel(peerId, event.channel);
		};

		this.peerConnections.set(peerId, peerConnection);
		return peerConnection;
	}

	// Close peer connection
	closePeerConnection(peerId) {
		const pc = this.peerConnections.get(peerId);
		if (pc) {
			pc.close();
			this.peerConnections.delete(peerId);
		}

		const dc = this.dataChannels.get(peerId);
		if (dc) {
			dc.close();
			this.dataChannels.delete(peerId);
		}
	}

	// Create offer
	async createOffer(peerId) {
		const peerConnection = this.getPeerConnection(peerId);
		try {
			const offer = await peerConnection.createOffer();
			await peerConnection.setLocalDescription(offer);
			return offer;
		} catch (error) {
			console.error('Error creating offer:', error);
			throw error;
		}
	}

	// Create answer
	async createAnswer(peerId, offer) {
		const peerConnection = this.getPeerConnection(peerId);
		try {
			await peerConnection.setRemoteDescription(new RTCSessionDescription(offer));
			const answer = await peerConnection.createAnswer();
			await peerConnection.setLocalDescription(answer);
			return answer;
		} catch (error) {
			console.error('Error creating answer:', error);
			throw error;
		}
	}

	// Add ICE candidate
	async addICECandidate(peerId, candidate) {
		const peerConnection = this.getPeerConnection(peerId);
		try {
			await peerConnection.addIceCandidate(new RTCIceCandidate(candidate));
		} catch (error) {
			console.error('Error adding ICE candidate:', error);
		}
	}

	// Setup data channel
	setupDataChannel(peerId, dataChannel) {
		this.dataChannels.set(peerId, dataChannel);

		dataChannel.onopen = () => {
			console.log(`Data channel opened with ${peerId}`);
		};

		dataChannel.onmessage = (event) => {
			this.onDataMessage(peerId, event.data);
		};

		dataChannel.onerror = (error) => {
			console.error(`Data channel error with ${peerId}:`, error);
		};

		dataChannel.onclose = () => {
			console.log(`Data channel closed with ${peerId}`);
			this.dataChannels.delete(peerId);
		};
	}

	// Send message via data channel
	sendMessage(peerId, message) {
		const dataChannel = this.dataChannels.get(peerId);
		if (dataChannel && dataChannel.readyState === 'open') {
			dataChannel.send(JSON.stringify(message));
		}
	}

	// Broadcast message to all peers
	broadcastMessage(message) {
		this.dataChannels.forEach((dc, peerId) => {
			if (dc.readyState === 'open') {
				dc.send(JSON.stringify(message));
			}
		});
	}

	// Get statistics
	async getStats(peerId) {
		const peerConnection = this.getPeerConnection(peerId);
		const stats = await peerConnection.getStats();
		const result = {
			inbound: {},
			outbound: {},
			candidate: {}
		};

		stats.forEach(report => {
			if (report.type === 'inbound-rtp') {
				result.inbound = {
					bytesReceived: report.bytesReceived,
					packetsLost: report.packetsLost,
					jitter: report.jitter
				};
			} else if (report.type === 'outbound-rtp') {
				result.outbound = {
					bytesSent: report.bytesSent,
					packetsLost: report.packetsLost,
					frameRate: report.framesPerSecond
				};
			} else if (report.type === 'candidate-pair' && report.state === 'succeeded') {
				result.candidate = {
					currentRoundTripTime: report.currentRoundTripTime,
					availableOutgoingBitrate: report.availableOutgoingBitrate
				};
			}
		});

		return result;
	}

	// Get peer connection
	getPeerConnection(peerId) {
		if (!this.peerConnections.has(peerId)) {
			return this.createPeerConnection(peerId);
		}
		return this.peerConnections.get(peerId);
	}

	// Callbacks - override these in the client
	onICECandidate(peerId, candidate) {
		console.log(`ICE candidate from ${peerId}:`, candidate);
	}

	onRemoteStream(peerId, stream) {
		console.log(`Remote stream from ${peerId}:`, stream);
	}

	onConnectionStateChange(peerId, state) {
		console.log(`Connection state change with ${peerId}:`, state);
	}

	onDataMessage(peerId, data) {
		console.log(`Data message from ${peerId}:`, data);
	}
}
