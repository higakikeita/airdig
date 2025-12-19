import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AlertTriangle, Filter, Search, ArrowUpDown, ArrowUp, ArrowDown, Clock, User, Activity } from 'lucide-react';
import apiClient from '../api/client';
import type { DriftEvent } from '../api/client';

type SortField = 'timestamp' | 'severity' | 'resource_type' | 'type';
type SortDirection = 'asc' | 'desc';

const DriftTable: React.FC = () => {
  const navigate = useNavigate();
  const [drifts, setDrifts] = useState<DriftEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [severityFilter, setSeverityFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [resourceTypeFilter, setResourceTypeFilter] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState<string>('');

  // Sort
  const [sortField, setSortField] = useState<SortField>('timestamp');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  // Pagination
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 20;

  useEffect(() => {
    loadDrifts();
    const interval = setInterval(() => {
      loadDrifts();
    }, 30000); // Update every 30 seconds
    return () => clearInterval(interval);
  }, [severityFilter, typeFilter, resourceTypeFilter]);

  const loadDrifts = async () => {
    try {
      setLoading(true);
      const params: any = { limit: 500 };
      if (severityFilter) params.severity = severityFilter;
      if (typeFilter) params.drift_type = typeFilter;
      if (resourceTypeFilter) params.resource_type = resourceTypeFilter;

      const response = await apiClient.listDrifts(params);
      setDrifts(response.drifts || []);
      setError(null);
    } catch (err) {
      console.error('Failed to load drifts:', err);
      setError('Failed to load drift events');
    } finally {
      setLoading(false);
    }
  };

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const getSortIcon = (field: SortField) => {
    if (sortField !== field) return <ArrowUpDown size={14} />;
    return sortDirection === 'asc' ? <ArrowUp size={14} /> : <ArrowDown size={14} />;
  };

  // Filter and sort drifts
  const filteredDrifts = drifts
    .filter(drift => {
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        return (
          drift.resource_id.toLowerCase().includes(query) ||
          drift.resource_type.toLowerCase().includes(query) ||
          drift.root_cause?.user_identity?.toLowerCase().includes(query)
        );
      }
      return true;
    })
    .sort((a, b) => {
      let aVal: any, bVal: any;

      switch (sortField) {
        case 'timestamp':
          aVal = new Date(a.timestamp).getTime();
          bVal = new Date(b.timestamp).getTime();
          break;
        case 'severity':
          const severityOrder = { critical: 4, high: 3, medium: 2, low: 1 };
          aVal = severityOrder[a.severity];
          bVal = severityOrder[b.severity];
          break;
        case 'resource_type':
          aVal = a.resource_type;
          bVal = b.resource_type;
          break;
        case 'type':
          aVal = a.type;
          bVal = b.type;
          break;
        default:
          return 0;
      }

      if (sortDirection === 'asc') {
        return aVal > bVal ? 1 : -1;
      } else {
        return aVal < bVal ? 1 : -1;
      }
    });

  // Pagination
  const totalPages = Math.ceil(filteredDrifts.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedDrifts = filteredDrifts.slice(startIndex, startIndex + itemsPerPage);

  const getSeverityColor = (severity: string) => {
    const colors: Record<string, string> = {
      critical: '#d13212',
      high: '#f89256',
      medium: '#f9b959',
      low: '#64748b',
    };
    return colors[severity] || colors.low;
  };

  const getTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      created: '#1d8102',
      modified: '#f9b959',
      deleted: '#d13212',
    };
    return colors[type] || '#64748b';
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (error) {
    return (
      <div style={{
        background: 'var(--color-background-container-content)',
        border: '1px solid var(--color-border-container-divider)',
        borderRadius: '8px',
        padding: 'var(--space-xl)',
        textAlign: 'center',
      }}>
        <AlertTriangle size={48} style={{ color: '#d13212', marginBottom: 'var(--space-m)' }} />
        <div style={{ fontSize: 'var(--font-size-heading-s)', color: 'white', marginBottom: 'var(--space-xs)' }}>
          Error Loading Drifts
        </div>
        <div style={{ fontSize: 'var(--font-size-body-m)', color: 'var(--color-text-body-secondary)' }}>
          {error}
        </div>
        <button
          onClick={loadDrifts}
          style={{
            marginTop: 'var(--space-m)',
            padding: 'var(--space-s) var(--space-m)',
            background: 'var(--color-background-button-primary-default)',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-l)', height: '100%' }}>
      {/* Filters & Search */}
      <div style={{
        background: 'var(--color-background-container-content)',
        border: '1px solid var(--color-border-container-divider)',
        borderRadius: '8px',
        padding: 'var(--space-l)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-m)', flexWrap: 'wrap' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-xs)' }}>
            <Search size={16} style={{ color: 'var(--color-text-body-secondary)' }} />
            <input
              type="text"
              placeholder="Search by resource ID, type, or user..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              style={{
                background: 'var(--color-background-container-header)',
                border: '1px solid var(--color-border-container-divider)',
                borderRadius: '4px',
                padding: 'var(--space-xs) var(--space-s)',
                color: 'white',
                fontSize: 'var(--font-size-body-m)',
                minWidth: '300px',
              }}
            />
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-xs)' }}>
            <Filter size={16} style={{ color: 'var(--color-text-body-secondary)' }} />
            <select
              value={severityFilter}
              onChange={(e) => setSeverityFilter(e.target.value)}
              style={{
                background: 'var(--color-background-container-header)',
                border: '1px solid var(--color-border-container-divider)',
                borderRadius: '4px',
                padding: 'var(--space-xs) var(--space-s)',
                color: 'white',
                fontSize: 'var(--font-size-body-m)',
              }}
            >
              <option value="">All Severities</option>
              <option value="critical">Critical</option>
              <option value="high">High</option>
              <option value="medium">Medium</option>
              <option value="low">Low</option>
            </select>
          </div>

          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            style={{
              background: 'var(--color-background-container-header)',
              border: '1px solid var(--color-border-container-divider)',
              borderRadius: '4px',
              padding: 'var(--space-xs) var(--space-s)',
              color: 'white',
              fontSize: 'var(--font-size-body-m)',
            }}
          >
            <option value="">All Types</option>
            <option value="created">Created</option>
            <option value="modified">Modified</option>
            <option value="deleted">Deleted</option>
          </select>

          <input
            type="text"
            placeholder="Filter by resource type..."
            value={resourceTypeFilter}
            onChange={(e) => setResourceTypeFilter(e.target.value)}
            style={{
              background: 'var(--color-background-container-header)',
              border: '1px solid var(--color-border-container-divider)',
              borderRadius: '4px',
              padding: 'var(--space-xs) var(--space-s)',
              color: 'white',
              fontSize: 'var(--font-size-body-m)',
              minWidth: '200px',
            }}
          />

          <div style={{ marginLeft: 'auto', color: 'var(--color-text-body-secondary)', fontSize: 'var(--font-size-body-s)' }}>
            {filteredDrifts.length} drift{filteredDrifts.length !== 1 ? 's' : ''} found
          </div>
        </div>
      </div>

      {/* Table */}
      <div style={{
        background: 'var(--color-background-container-content)',
        border: '1px solid var(--color-border-container-divider)',
        borderRadius: '8px',
        overflow: 'hidden',
        flex: 1,
      }}>
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '400px', color: 'white' }}>
            <Activity size={24} style={{ animation: 'spin 1s linear infinite' }} />
            <span style={{ marginLeft: 'var(--space-s)' }}>Loading drifts...</span>
          </div>
        ) : paginatedDrifts.length === 0 ? (
          <div style={{ textAlign: 'center', padding: 'var(--space-xxl)', color: 'var(--color-text-body-secondary)' }}>
            <AlertTriangle size={48} style={{ marginBottom: 'var(--space-m)' }} />
            <div style={{ fontSize: 'var(--font-size-heading-s)', marginBottom: 'var(--space-xs)' }}>
              No drifts found
            </div>
            <div style={{ fontSize: 'var(--font-size-body-m)' }}>
              Try adjusting your filters or search query
            </div>
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ background: 'var(--color-background-container-header)', borderBottom: '2px solid var(--color-border-container-divider)' }}>
                  <th style={headerCellStyle} onClick={() => handleSort('timestamp')}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-xs)', cursor: 'pointer' }}>
                      <Clock size={14} />
                      Timestamp
                      {getSortIcon('timestamp')}
                    </div>
                  </th>
                  <th style={headerCellStyle} onClick={() => handleSort('severity')}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-xs)', cursor: 'pointer' }}>
                      <AlertTriangle size={14} />
                      Severity
                      {getSortIcon('severity')}
                    </div>
                  </th>
                  <th style={headerCellStyle} onClick={() => handleSort('type')}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-xs)', cursor: 'pointer' }}>
                      Type
                      {getSortIcon('type')}
                    </div>
                  </th>
                  <th style={headerCellStyle} onClick={() => handleSort('resource_type')}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-xs)', cursor: 'pointer' }}>
                      Resource Type
                      {getSortIcon('resource_type')}
                    </div>
                  </th>
                  <th style={headerCellStyle}>Resource ID</th>
                  <th style={headerCellStyle}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-xs)' }}>
                      <User size={14} />
                      User
                    </div>
                  </th>
                  <th style={headerCellStyle}>Event</th>
                </tr>
              </thead>
              <tbody>
                {paginatedDrifts.map((drift, index) => (
                  <tr
                    key={drift.id}
                    onClick={() => navigate(`/drift/${drift.id}`)}
                    style={{
                      borderBottom: '1px solid var(--color-border-container-divider)',
                      transition: 'background 0.15s ease',
                      cursor: 'pointer',
                      background: index % 2 === 0 ? 'transparent' : 'var(--color-background-container-header)',
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.background = 'var(--color-background-button-primary-hover)';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.background = index % 2 === 0 ? 'transparent' : 'var(--color-background-container-header)';
                    }}
                  >
                    <td style={cellStyle}>{formatTimestamp(drift.timestamp)}</td>
                    <td style={cellStyle}>
                      <span style={{
                        padding: '2px 8px',
                        borderRadius: '4px',
                        fontSize: 'var(--font-size-body-s)',
                        fontWeight: '600',
                        background: getSeverityColor(drift.severity) + '33',
                        color: getSeverityColor(drift.severity),
                      }}>
                        {drift.severity.toUpperCase()}
                      </span>
                    </td>
                    <td style={cellStyle}>
                      <span style={{
                        padding: '2px 8px',
                        borderRadius: '4px',
                        fontSize: 'var(--font-size-body-s)',
                        fontWeight: '600',
                        background: getTypeColor(drift.type) + '33',
                        color: getTypeColor(drift.type),
                      }}>
                        {drift.type}
                      </span>
                    </td>
                    <td style={cellStyle}>
                      <code style={{ fontSize: 'var(--font-size-body-s)', color: 'var(--aws-smile-orange)' }}>
                        {drift.resource_type}
                      </code>
                    </td>
                    <td style={cellStyle}>
                      <code style={{ fontSize: 'var(--font-size-body-s)', color: 'var(--color-text-body-secondary)' }}>
                        {drift.resource_id}
                      </code>
                    </td>
                    <td style={cellStyle}>
                      {drift.root_cause?.user_identity || '-'}
                    </td>
                    <td style={cellStyle}>
                      {drift.root_cause?.event_name || '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          gap: 'var(--space-m)',
          padding: 'var(--space-m)',
        }}>
          <button
            onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
            disabled={currentPage === 1}
            style={{
              ...paginationButtonStyle,
              opacity: currentPage === 1 ? 0.5 : 1,
              cursor: currentPage === 1 ? 'not-allowed' : 'pointer',
            }}
          >
            Previous
          </button>

          <span style={{ color: 'white', fontSize: 'var(--font-size-body-m)' }}>
            Page {currentPage} of {totalPages}
          </span>

          <button
            onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
            disabled={currentPage === totalPages}
            style={{
              ...paginationButtonStyle,
              opacity: currentPage === totalPages ? 0.5 : 1,
              cursor: currentPage === totalPages ? 'not-allowed' : 'pointer',
            }}
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
};

const headerCellStyle: React.CSSProperties = {
  padding: 'var(--space-m)',
  textAlign: 'left',
  fontSize: 'var(--font-size-body-s)',
  fontWeight: '700',
  color: 'var(--color-text-label)',
  textTransform: 'uppercase',
  letterSpacing: '0.5px',
};

const cellStyle: React.CSSProperties = {
  padding: 'var(--space-m)',
  fontSize: 'var(--font-size-body-m)',
  color: 'white',
};

const paginationButtonStyle: React.CSSProperties = {
  padding: 'var(--space-xs) var(--space-m)',
  background: 'var(--color-background-button-primary-default)',
  border: 'none',
  borderRadius: '4px',
  color: 'white',
  fontSize: 'var(--font-size-body-m)',
  fontWeight: '500',
};

export default DriftTable;
