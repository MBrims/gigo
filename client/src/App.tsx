import { BrowserRouter, Route, Routes } from 'react-router';
import './App.css';
import Download from './pages/download';
import Home from './pages/home';

const App = () => {
  // const params = new URLSearchParams(window.location.search);
  // const fileParam = params.get("file");

// {/* <div className="content">
//       {fileParam ? <Download /> : <Home />}
//     </div>
//     </> */}

  return (
    <BrowserRouter>
      <Routes>
        <Route path ="/" element={<Home />} />
        <Route path='/download' element={<Download />} />
      </Routes>
    </BrowserRouter>
  );
};

export default App;
