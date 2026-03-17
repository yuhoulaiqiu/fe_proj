import { Navigate, Route, Routes } from 'react-router-dom'
import './App.css'
import AdminLayout from './components/AdminLayout.jsx'
import RequireAdmin from './components/RequireAdmin.jsx'
import SiteLayout from './components/SiteLayout.jsx'
import ToastProvider from './components/ui/Toast.jsx'
import ActivitiesPage from './pages/ActivitiesPage.jsx'
import ActivityDetailPage from './pages/ActivityDetailPage.jsx'
import AdminLoginPage from './pages/admin/AdminLoginPage.jsx'
import AdminLostItemFormPage from './pages/admin/AdminLostItemFormPage.jsx'
import AdminLostItemsPage from './pages/admin/AdminLostItemsPage.jsx'
import HomePage from './pages/HomePage.jsx'
import LostFoundPage from './pages/LostFoundPage.jsx'
import LostItemDetailPage from './pages/LostItemDetailPage.jsx'
import NotFoundPage from './pages/NotFoundPage.jsx'
import ServicesPage from './pages/ServicesPage.jsx'
import ServiceDetailPage from './pages/ServiceDetailPage.jsx'

function App() {
  return (
    <ToastProvider>
      <Routes>
        <Route element={<SiteLayout />}>
          <Route index element={<HomePage />} />
          <Route path="/activities" element={<ActivitiesPage />} />
          <Route path="/activities/:id" element={<ActivityDetailPage />} />
          <Route path="/services" element={<ServicesPage />} />
          <Route path="/services/:id" element={<ServiceDetailPage />} />
          <Route path="/lost-found" element={<LostFoundPage />} />
          <Route path="/lost-found/:id" element={<LostItemDetailPage />} />
        </Route>

        <Route path="/admin/login" element={<AdminLoginPage />} />
        <Route
          path="/admin"
          element={
            <RequireAdmin>
              <AdminLayout />
            </RequireAdmin>
          }
        >
          <Route index element={<Navigate to="/admin/lost-items" replace />} />
          <Route path="lost-items" element={<AdminLostItemsPage />} />
          <Route path="lost-items/new" element={<AdminLostItemFormPage />} />
          <Route path="lost-items/:id/edit" element={<AdminLostItemFormPage />} />
        </Route>

        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </ToastProvider>
  )
}

export default App
