import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { User, Mail, Phone, Calendar, Shield, Save } from 'lucide-react';
import { useAuth } from '@/contexts/AuthContext';
import { useProfile } from '@/hooks/useProfile';

interface ProfileFormData {
  full_name: string;
  gmail?: string;
  phone?: string;
}

const Profile: React.FC = () => {
  const { user } = useAuth();
  const { profile, isLoading, updateProfile } = useProfile();
  const [saving, setSaving] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState<ProfileFormData>({
    full_name: '',
    gmail: '',
    phone: '',
  });
  const [alert, setAlert] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  useEffect(() => {
    if (profile) {
      setFormData({
        full_name: profile.full_name || '',
        gmail: profile.gmail || '',
        phone: profile.phone || '',
      });
    }
  }, [profile]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSave = async () => {
    setSaving(true);
    setAlert(null);

    try {
      const result = await updateProfile({
        full_name: formData.full_name,
        gmail: formData.gmail || null,
        phone: formData.phone || null,
      });

      if (result.success) {
        setIsEditing(false);
        setAlert({ type: 'success', message: 'Profile updated successfully!' });
      } else {
        throw new Error(result.error || 'Failed to update profile');
      }
    } catch (error) {
      console.error('Error updating profile:', error);
      setAlert({ type: 'error', message: error instanceof Error ? error.message : 'Failed to update profile' });
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    if (profile) {
      setFormData({
        full_name: profile.full_name || '',
        gmail: profile.gmail || '',
        phone: profile.phone || '',
      });
    }
    setIsEditing(false);
    setAlert(null);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'trial':
        return 'bg-blue-100 text-blue-800';
      case 'expired':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (!profile || !user) {
    return (
      <div className="flex items-center justify-center h-96">
        <Alert>
          <AlertDescription>Failed to load profile data. Please try refreshing the page.</AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 max-w-4xl">
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Profile</h1>
            <p className="text-gray-600">Manage your account settings and preferences</p>
          </div>
          {!isEditing && (
            <Button onClick={() => setIsEditing(true)} className="flex items-center gap-2">
              <User className="h-4 w-4" />
              Edit Profile
            </Button>
          )}
        </div>

        {/* Alert */}
        {alert && (
          <Alert className={alert.type === 'error' ? 'border-red-500 bg-red-50' : 'border-green-500 bg-green-50'}>
            <AlertDescription className={alert.type === 'error' ? 'text-red-700' : 'text-green-700'}>
              {alert.message}
            </AlertDescription>
          </Alert>
        )}

        <div className="grid gap-6 md:grid-cols-2">
          {/* Profile Information */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <User className="h-5 w-5" />
                Profile Information
              </CardTitle>
              <CardDescription>
                Your personal information and contact details
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="full_name">Full Name</Label>
                <Input
                  id="full_name"
                  name="full_name"
                  value={formData.full_name}
                  onChange={handleInputChange}
                  disabled={!isEditing}
                  placeholder="Enter your full name"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <div className="flex items-center p-3 bg-gray-50 border border-gray-200 rounded-md">
                  <Mail className="h-4 w-4 text-gray-400 mr-2" />
                  <span className="text-sm text-gray-700 select-text">{user.email}</span>
                </div>
                <p className="text-sm text-gray-500">Email cannot be changed</p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="gmail">Gmail (Optional)</Label>
                <Input
                  id="gmail"
                  name="gmail"
                  type="email"
                  value={formData.gmail}
                  onChange={handleInputChange}
                  disabled={!isEditing}
                  placeholder="Enter your Gmail address"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="phone">Phone Number (Optional)</Label>
                <Input
                  id="phone"
                  name="phone"
                  value={formData.phone}
                  onChange={handleInputChange}
                  disabled={!isEditing}
                  placeholder="Enter your phone number"
                />
              </div>
            </CardContent>
          </Card>

          {/* Account Status */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Shield className="h-5 w-5" />
                Account Status
              </CardTitle>
              <CardDescription>
                Your account information and status
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">Status:</span>
                <Badge className={getStatusColor(profile.status)}>
                  {profile.status}
                </Badge>
              </div>

              {profile.expired && (
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Expires:</span>
                  <span className="text-sm text-gray-600">{formatDate(profile.expired)}</span>
                </div>
              )}

              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">Member Since:</span>
                <span className="text-sm text-gray-600">{formatDate(profile.created_at)}</span>
              </div>

              {profile.last_login && (
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Last Login:</span>
                  <span className="text-sm text-gray-600">{formatDate(profile.last_login)}</span>
                </div>
              )}

              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">Account ID:</span>
                <span className="text-sm text-gray-600 font-mono">{user.id.substring(0, 8)}...</span>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Action Buttons */}
        {isEditing && (
          <div className="flex justify-end gap-3">
            <Button variant="outline" onClick={handleCancel}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={saving} className="flex items-center gap-2">
              <Save className="h-4 w-4" />
              {saving ? 'Saving...' : 'Save Changes'}
            </Button>
          </div>
        )}
      </div>
    </div>
  );
};

export default Profile;
