import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from './AuthContext';

export function RegisterPage() {
  const { sendCode, verifyCode, register } = useAuth();
  const navigate = useNavigate();
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [loginValue, setLoginValue] = useState('');
  const [code, setCode] = useState('');
  const [debugCode, setDebugCode] = useState<string | null>(null);
  const [confirmToken, setConfirmToken] = useState('');
  const [username, setUsername] = useState('');
  const [fullName, setFullName] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);

  async function send(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      const dbg = await sendCode(loginValue);
      setDebugCode(dbg);
      setStep(2);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed');
    }
  }

  async function verify(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      const token = await verifyCode(loginValue, code);
      setConfirmToken(token);
      setStep(3);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed');
    }
  }

  async function finish(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await register(confirmToken, username, password, fullName || undefined);
      navigate('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed');
    }
  }

  return (
    <div className="auth-card">
      <h1>Create account</h1>
      {step === 1 && (
        <form onSubmit={send}>
          <label>
            Email or phone
            <input value={loginValue} onChange={(e) => setLoginValue(e.target.value)} required />
          </label>
          <button type="submit">Send code</button>
        </form>
      )}
      {step === 2 && (
        <form onSubmit={verify}>
          {debugCode && <p className="muted">Dev OTP: <strong>{debugCode}</strong></p>}
          <label>
            Code
            <input value={code} onChange={(e) => setCode(e.target.value)} required />
          </label>
          <button type="submit">Verify</button>
        </form>
      )}
      {step === 3 && (
        <form onSubmit={finish}>
          <label>
            Username
            <input value={username} onChange={(e) => setUsername(e.target.value)} required />
          </label>
          <label>
            Display name
            <input value={fullName} onChange={(e) => setFullName(e.target.value)} />
          </label>
          <label>
            Password
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
          </label>
          <button type="submit">Register</button>
        </form>
      )}
      {error && <p className="error">{error}</p>}
      <p className="muted"><Link to="/login">Back to login</Link></p>
    </div>
  );
}
