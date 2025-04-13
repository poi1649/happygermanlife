// Example WebSocket client for testing the Speech-to-Text streaming API

// Configuration
const SERVER_URL = `ws://localhost:8080/api/speech?Username=${encodeURIComponent('test_user')}`;
const USERNAME = 'test_user';

// Sample audio recording function (simplified)
function startRecording() {
  console.log('Started recording...');
  
  // This is a placeholder for actual audio recording logic
  // In a real application, you would:
  // 1. Use the Web Audio API or similar to capture audio
  // 2. Convert the audio to the correct format (LINEAR16, 16kHz)
  // 3. Send the audio data to the server
  
  // Example of what sending audio data would look like:
  // audioRecorder.ondataavailable = (event) => {
  //   if (socket.readyState === WebSocket.OPEN) {
  //     socket.send(event.data);
  //   }
  // };
}

// WebSocket connection
function connectToServer() {
  const socket = new WebSocket(SERVER_URL);
  
  // Add username as a query parameter
  // WebSocket URL should already include the Username query parameter
  socket.addEventListener('open', (event) => {
    console.log('Connected to server');
    
    // Start recording and sending audio data
    startRecording();
    
    // Simulate sending audio data
    console.log('Simulating sending audio data...');
    
    // Example: Send a dummy message every second for 5 seconds
    let count = 0;
    const interval = setInterval(() => {
      if (count < 5) {
        // In a real application, this would be actual audio data
        const dummyAudioData = new Uint8Array([0, 1, 2, 3, 4]);
        socket.send(dummyAudioData);
        console.log(`Sent audio data chunk ${count + 1}`);
        count++;
      } else {
        clearInterval(interval);
        console.log('Finished sending audio data');
        
        // Close the connection after a delay
        setTimeout(() => {
          socket.close();
          console.log('Connection closed');
        }, 1000);
      }
    }, 1000);
  });
  
  socket.addEventListener('message', (event) => {
    console.log('Received message from server:', event.data);
    
    try {
      const data = JSON.parse(event.data);
      console.log('Transcription:', data.transcript);
      console.log('Is final:', data.final);
    } catch (error) {
      console.error('Error parsing server message:', error);
    }
  });
  
  socket.addEventListener('close', (event) => {
    console.log('Connection closed with code:', event.code);
  });
  
  socket.addEventListener('error', (event) => {
    console.error('WebSocket error:', event);
  });
}

// Example of how to test the response generation API
async function testResponseGeneration() {
  try {
    const response = await fetch('http://localhost:8080/api/generate-response', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: USERNAME,
        context: {
          service: 'internet',
          issue: 'connection problem',
        },
      }),
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    console.log('Generated response:', data);
  } catch (error) {
    console.error('Error testing response generation:', error);
  }
}

// Start the demo
console.log('Starting WebSocket client demo...');
// Uncomment the line below to test the WebSocket connection
// connectToServer();

// To test the response generation after a conversation has been recorded
// Uncomment the line below
// testResponseGeneration();
