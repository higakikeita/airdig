import './App.css'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import CloudscapeDashboard from './components/CloudscapeDashboard'
import DriftDetail from './components/DriftDetail'
import ResourcesPage from './components/ResourcesPage'

function App() {
  return (
    <Router basename="/ui">
      <div className="App" style={{ margin: 0, padding: 0 }}>
        <Routes>
          <Route path="/" element={<Layout><CloudscapeDashboard /></Layout>} />
          <Route path="/resources" element={<Layout><ResourcesPage /></Layout>} />
          <Route path="/drift/:id" element={<Layout><DriftDetail /></Layout>} />
        </Routes>
      </div>
    </Router>
  )
}

export default App
