// Client application logic
const webrtc = new WebRTCManager();
let ws = null;
let currentRoomID = null;
let currentUserID = null;
let localVideoElement = null;
let audioEnabled = true;
let videoEnabled = true;

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
	localVideoElement = document.getElementById('localVideo');
	updateStatus('Disconnected');
});

// Join room
async function joinRoom() {
	const roomInput = document.getElementById('roomInput');
	const userInput = document.getElementById('userInput');

	currentRoomID = roomInput.value || 'default-room';
	currentUserID = userInput.value || 'user-' + Date.now();

	if (!currentRoomID || !currentUserID) {
		alert('Please enter room ID and user name');
		return;
	}

	try {
		// Get local stream
		await webrtc.getLocalStream(audioEnabled, videoEnabled);
		if (localVideoElement && webrtc.localStream) {
			localVideoElement.srcObject = webrtc.localStream;
		}

		// Connect to WebSocket
		connectWebSocket();

		// Update UI
		document.getElementById('joinBtn').disabled = true;
		document.getElementById('leaveBtn').disabled = false;
		updateStatus('Connecting');
	} catch (error) {
		console.error('Error joining room:', error);
		alert('Failed to access camera/microphone: ' + error.message);
	}
}

// Leave room
function leaveRoom() {
	if (ws) {
		ws.send(JSON.stringify({
			type: 'leave_room',
			room_id: currentRoomID
		}));
		ws.close();
	}

	webrtc.stopLocalStream();
	webrtc.peerConnections.forEach((pc, peerId) => {
		webrtc.closePeerConnection(peerId);
		removeRemoteVideo(peerId);
	});

	currentRoomID = null;
	currentUserID = null;

	document.getElementById('joinBtn').disabled = false;
	document.getElementById('leaveBtn').disabled = true;
	document.getElementById('peerList').innerHTML = '';
	document.getElementById('peerCount').textContent = '0';
	document.getElementById('videoGrid').innerHTML = '<div class="video-container local"><video id="localVideo" autoplay playsinline muted></video><span class="video-label">You</span></div>';
	localVideoElement = document.getElementById('localVideo');

	updateStatus('Disconnected');
}

// Connect to WebSocket
function connectWebSocket() {
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const wsURL = `${protocol}//${window.location.host}/ws`;

	ws = new WebSocket(wsURL);

	ws.onopen = () => {
		console.log('WebSocket connected');
		updateStatus('Connecting');

		// Send join room message
		ws.send(JSON.stringify({
			type: 'join_room',
			room_id: currentRoomID,
			user_id: currentUserID
		}));
	};

	ws.onmessage = (event) => {
		handleWebSocketMessage(JSON.parse(event.data));
	};

	ws.onerror = (error) => {
		console.error('WebSocket error:', error);
		updateStatus('Error');
	};

	ws.onclose = () => {
		console.log('WebSocket disconnected');
		updateStatus('Disconnected');
	};
}

// Handle WebSocket messages
async function handleWebSocketMessage(msg) {
	switch (msg.type) {
		case 'joined_room':
			handleRoomJoined(msg);
			break;
		case 'peer_joined':
			handlePeerJoined(msg);
			break;
		case 'peer_left':
			handlePeerLeft(msg);
			break;
		case 'offer':
			handleOffer(msg);
			break;
		case 'answer':
			handleAnswer(msg);
			break;
		case 'ice_candidate':
			handleICECandidate(msg);
			break;
		case 'message':
			handleMessage(msg);
			break;
		case 'error':
			console.error('Server error:', msg.error);
			alert('Error: ' + msg.error);
			break;
	}
}

// Handle room joined
async function handleRoomJoined(msg) {
	updateStatus('Connected');
	console.log('Joined room:', msg.room_id);
	updatePeerList(msg.peers || []);
	addSystemMessage(`You joined room ${msg.room_id}`);

	// Create offer to existing peers
	if (msg.peers && msg.peers.length > 0) {
		for (const peer of msg.peers) {
			const offer = await webrtc.createOffer(peer.id);
			ws.send(JSON.stringify({
				type: 'offer',
				peer_id: peer.id,
				sdp: offer.sdp
			}));
		}
	}
}

// Handle peer joined
async function handlePeerJoined(msg) {
	console.log('Peer joined:', msg.peer_id);
	updatePeerList([{ id: msg.peer_id, user: msg.user_id }]);
	addSystemMessage(`${msg.user_id} joined the room`);
}

// Handle peer left
function handlePeerLeft(msg) {
	console.log('Peer left:', msg.peer_id);
	removeRemoteVideo(msg.peer_id);
	webrtc.closePeerConnection(msg.peer_id);
	addSystemMessage(`${msg.user_id} left the room`);
}

// Handle offer
async function handleOffer(msg) {
	const offer = new RTCSessionDescription({
		type: 'offer',
		sdp: msg.sdp
	});

	const answer = await webrtc.createAnswer(msg.peer_id, offer);
	ws.send(JSON.stringify({
		type: 'answer',
		peer_id: msg.peer_id,
		sdp: answer.sdp
	}));
}

// Handle answer
async function handleAnswer(msg) {
	const answer = new RTCSessionDescription({
		type: 'answer',
		sdp: msg.sdp
	});

	const pc = webrtc.getPeerConnection(msg.peer_id);
	await pc.setRemoteDescription(answer);
}

// Handle ICE candidate
async function handleICECandidate(msg) {
	const candidate = new RTCIceCandidate(msg.candidate);
	await webrtc.addICECandidate(msg.peer_id, candidate);
}

// Handle message
function handleMessage(msg) {
	if (msg.message) {
		addChatMessage(msg.message.sender_id, msg.message.content, false);
	}
}

// Override WebRTC callbacks
webrtc.onRemoteStream = (peerId, stream) => {
	console.log('Remote stream:', peerId);
	displayRemoteVideo(peerId, stream);
};

webrtc.onConnectionStateChange = (peerId, state) => {
	console.log(`Connection state with ${peerId}:`, state);
	if (state === 'connected') {
		addSystemMessage(`Connected to ${peerId}`);
	}
};

webrtc.onICECandidate = (peerId, candidate) => {
	ws.send(JSON.stringify({
		type: 'ice_candidate',
		peer_id: peerId,
		candidate: {
			candidate: candidate.candidate,
			sdp_mline_index: candidate.sdpMLineIndex,
			sdp_mid: candidate.sdpMid
		}
	}));
};

// Display remote video
function displayRemoteVideo(peerId, stream) {
	let videoContainer = document.getElementById(`video-${peerId}`);
	if (!videoContainer) {
		const grid = document.getElementById('videoGrid');
		videoContainer = document.createElement('div');
		videoContainer.id = `video-${peerId}`;
		videoContainer.className = 'video-container remote';
		videoContainer.innerHTML = `
			<video autoplay playsinline></video>
			<span class="video-label">${peerId.substring(0, 8)}...</span>
		`;
		grid.appendChild(videoContainer);
	}

	const video = videoContainer.querySelector('video');
	video.srcObject = stream;
}

// Remove remote video
function removeRemoteVideo(peerId) {
	const container = document.getElementById(`video-${peerId}`);
	if (container) {
		container.remove();
	}
}

// Update peer list
function updatePeerList(peers) {
	const peerList = document.getElementById('peerList');
	peerList.innerHTML = '';

	peers.forEach(peer => {
		const item = document.createElement('div');
		item.className = 'peer-item';
		item.textContent = peer.user || peer.id;
		peerList.appendChild(item);
	});

	document.getElementById('peerCount').textContent = peers.length;
}

// Send message
function sendMessage() {
	const input = document.getElementById('messageInput');
	const content = input.value.trim();

	if (!content) return;

	ws.send(JSON.stringify({
		type: 'message',
		content: content
	}));

	addChatMessage(currentUserID, content, true);
	input.value = '';
}

// Add chat message
function addChatMessage(sender, content, isOwn) {
	const messageList = document.getElementById('messageList');
	const item = document.createElement('div');
	item.className = `message-item ${isOwn ? 'own' : ''}`;

	const time = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

	item.innerHTML = `
		<div class="message-sender">${sender}</div>
		<div class="message-content">${escapeHtml(content)}</div>
		<div class="message-time">${time}</div>
	`;

	messageList.appendChild(item);
	messageList.scrollTop = messageList.scrollHeight;
}

// Add system message
function addSystemMessage(content) {
	const messageList = document.getElementById('messageList');
	const item = document.createElement('div');
	item.className = 'message-item system';

	item.innerHTML = `<div class="message-content">${content}</div>`;

	messageList.appendChild(item);
	messageList.scrollTop = messageList.scrollHeight;
}

// Toggle audio
function toggleAudio() {
	audioEnabled = !audioEnabled;
	if (webrtc.localStream) {
		webrtc.localStream.getAudioTracks().forEach(track => {
			track.enabled = audioEnabled;
		});
	}

	const btn = document.getElementById('audioBtn');
	btn.textContent = audioEnabled ? '📢 Audio ON' : '🔇 Audio OFF';
	btn.style.opacity = audioEnabled ? '1' : '0.5';
}

// Toggle video
function toggleVideo() {
	videoEnabled = !videoEnabled;
	if (webrtc.localStream) {
		webrtc.localStream.getVideoTracks().forEach(track => {
			track.enabled = videoEnabled;
		});
	}

	const btn = document.getElementById('videoBtn');
	btn.textContent = videoEnabled ? '📹 Video ON' : '📹 Video OFF';
	btn.style.opacity = videoEnabled ? '1' : '0.5';
}

// Toggle screen share (placeholder)
function toggleScreen() {
	alert('Screen sharing coming soon!');
}

// Handle Enter key in message input
function handleKeypress(event) {
	if (event.key === 'Enter') {
		sendMessage();
	}
}

// Update status
function updateStatus(status) {
	const statusEl = document.getElementById('status');
	statusEl.textContent = status;

	if (status === 'Connected') {
		statusEl.className = 'status-badge connected';
	} else if (status === 'Connecting') {
		statusEl.className = 'status-badge connecting';
	} else {
		statusEl.className = 'status-badge disconnected';
	}
}

// Escape HTML for security
function escapeHtml(text) {
	const map = {
		'&': '&amp;',
		'<': '&lt;',
		'>': '&gt;',
		'"': '&quot;',
		"'": '&#039;'
	};
	return text.replace(/[&<>"']/g, m => map[m]);
}

// Update statistics periodically
setInterval(async () => {
	if (webrtc.peerConnections.size > 0) {
		try {
			const peerId = Array.from(webrtc.peerConnections.keys())[0];
			const stats = await webrtc.getStats(peerId);

			if (stats.candidate.currentRoundTripTime !== undefined) {
				document.getElementById('latency').textContent =
					(stats.candidate.currentRoundTripTime * 1000).toFixed(0) + ' ms';
			}

			if (stats.inbound.packetsLost !== undefined) {
				document.getElementById('packetsLost').textContent = stats.inbound.packetsLost;
			}

			if (stats.candidate.availableOutgoingBitrate !== undefined) {
				const bandwidth = (stats.candidate.availableOutgoingBitrate / 1000000).toFixed(2);
				document.getElementById('bandwidth').textContent = bandwidth + ' Mbps';
			}
		} catch (error) {
			console.log('Stats error:', error);
		}
	}
}, 1000);
