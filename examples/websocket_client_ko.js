// WebSocket 클라이언트 예제 (한국어 주석)

// 설정
const USERNAME = 'test_user';
const SERVER_URL = `ws://localhost:8080/api/speech?Username=${encodeURIComponent(USERNAME)}`;

// WebSocket 연결
function connectToServer() {
  console.log('서버에 연결 시도 중...');
  
  // WebSocket 연결 생성 (Username을 쿼리 파라미터로 전달)
  const socket = new WebSocket(SERVER_URL);
  
  // 연결 성공 이벤트 핸들러
  socket.addEventListener('open', (event) => {
    console.log('서버에 연결되었습니다');
    
    // 오디오 녹음 시뮬레이션 시작
    console.log('오디오 데이터 전송 시뮬레이션 시작...');
    
    // 5초 동안 1초마다 더미 오디오 데이터 전송
    let count = 0;
    const interval = setInterval(() => {
      if (count < 5) {
        // 실제 애플리케이션에서는 실제 오디오 데이터가 됨
        const dummyAudioData = new Uint8Array([0, 1, 2, 3, 4]);
        socket.send(dummyAudioData);
        console.log(`오디오 데이터 청크 ${count + 1} 전송 완료`);
        count++;
      } else {
        clearInterval(interval);
        console.log('오디오 데이터 전송 완료');
        
        // 1초 후 연결 종료
        setTimeout(() => {
          socket.close();
          console.log('연결 종료됨');
        }, 1000);
      }
    }, 1000);
  });
  
  // 메시지 수신 이벤트 핸들러
  socket.addEventListener('message', (event) => {
    console.log('서버로부터 메시지 수신:', event.data);
    
    try {
      const data = JSON.parse(event.data);
      console.log('텍스트 변환 결과:', data.transcript);
      console.log('최종 결과 여부:', data.final);
    } catch (error) {
      console.error('서버 메시지 파싱 오류:', error);
    }
  });
  
  // 연결 종료 이벤트 핸들러
  socket.addEventListener('close', (event) => {
    console.log('연결 종료 코드:', event.code);
  });
  
  // 오류 이벤트 핸들러
  socket.addEventListener('error', (event) => {
    console.error('WebSocket 오류:', event);
  });
}

// 응답 생성 API 테스트 함수
async function testResponseGeneration() {
  console.log('응답 생성 API 테스트 중...');
  
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
          issue: '연결 문제',
        },
      }),
    });
    
    if (!response.ok) {
      throw new Error(`HTTP 오류! 상태: ${response.status}`);
    }
    
    const data = await response.json();
    console.log('생성된 응답:', data);
  } catch (error) {
    console.error('응답 생성 테스트 오류:', error);
  }
}

// 테스트 시작
console.log('WebSocket 클라이언트 데모 시작...');

// WebSocket 연결 테스트
// connectToServer();

// 응답 생성 테스트
// testResponseGeneration();
