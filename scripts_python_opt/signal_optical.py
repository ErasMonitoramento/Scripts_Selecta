import argparse
from pysnmp.hlapi import *

def snmp_get_single(oid, host, community='public', port=161):
    result = []
    for (errorIndication,
         errorStatus,
         errorIndex,
         varBinds) in getCmd(SnmpEngine(),
                             CommunityData(community),
                             UdpTransportTarget((host, port)),
                             ContextData(),
                             ObjectType(ObjectIdentity(oid))):

        if errorIndication:
            print(f"Error Indication: {errorIndication}")
            break
        elif errorStatus:
            print(f"Error Status: {errorStatus.prettyPrint()} at {errorIndex and varBinds[int(errorIndex) - 1][0] or '?'}")
            break
        else:
            for varBind in varBinds:
                result.append(' = '.join([x.prettyPrint() for x in varBind]))
    return result

def process_snmp_data(snmp_data, lane_type):
    processed_data = []
    for item in snmp_data:
        value = item.split(' = ')[-1]
        if value and value != '""':
            if ',' in value:  # Multi Lane
                if lane_type == 'M':
                    values = value.split(',')
                    processed_values = []
                    for v in values[:4]:  # Only take the first four values
                        try:
                            processed_values.append(int(float(v)))  # Convert to float, then round to int
                        except ValueError:
                            processed_values.append(float('nan'))
                    processed_data.append(processed_values)
            else:  # Single Lane
                if lane_type == 'S':
                    try:
                        processed_values = [int(float(value))]  # Convert to float, then round to int
                    except ValueError:
                        processed_values = [float('nan')]
                    processed_data.append(processed_values)
    return processed_data

def main():
    parser = argparse.ArgumentParser(description='Get SNMP optical signal data.')
    parser.add_argument('community', type=str, help='SNMP community')
    parser.add_argument('host', type=str, help='SNMP host')
    parser.add_argument('index', type=str, help='SNMP index for the port')
    parser.add_argument('port', type=int, help='SNMP port', default=161, nargs='?')
    parser.add_argument('signal_type', type=str, choices=['RX', 'TX'], help='Type of signal (RX or TX)')
    parser.add_argument('lane_type', type=str, choices=['M', 'S'], help='Type of lane (M for Multi Lane, S for Single Lane)')

    args = parser.parse_args()

    # OIDs
    if args.signal_type == 'RX':
        oid = f'1.3.6.1.4.1.2011.5.25.31.1.1.3.1.32.{args.index}'
    elif args.signal_type == 'TX':
        oid = f'1.3.6.1.4.1.2011.5.25.31.1.1.3.1.33.{args.index}'

    # Get SNMP data
    data = snmp_get_single(oid, args.host, args.community, args.port)

    # Process SNMP data
    processed_data = process_snmp_data(data, args.lane_type)

    # Extract the first value from each list of processed data
    first_value = processed_data[0][0] if processed_data else None

    if first_value is not None:
        print(first_value)

if __name__ == "__main__":
    main()
