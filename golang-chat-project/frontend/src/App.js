import React, { Component } from 'react';
import Header from './components/Header/Header';
import ChatHistory from './components/ChatHistory/ChatHistory';
import ChatInput from './components/ChatInput/ChatInput';
import './App.css';
import {connect, sendMsg} from './api';

class App extends Component {
  constructor(props){
    super(props);
    this.state = {
      ChatHistory: []
    }
  }

  componentDidMount() {
    connect((msg) => {
      console.log("New Message")
      this.setState(prevState => ({
        ChatHistory: [...prevState.ChatHistory, msg]
      }))
      console.log(this.state)
    })
  }

  send(event) {
    if (event.keyCode === 13) {
      sendMsg(event.target.value);
      event.target.value = "";
    }
  }
  
  render(){
    return(
      <div className='App'>
        <Header/>
        <ChatHistory chatHistory={this.state.ChatHistory}/>
        <ChatInput send={this.send}/>
      </div>
    )
  }
}

export default App;
