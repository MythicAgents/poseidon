from mythic_payloadtype_container.MythicCommandBase import *
import sys


class XpcArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="list",
                type=ParameterType.Boolean,
                description="Flag to indicate asking launchd to list running services",
                default_value=True,
                parameter_group_info=[
                    ParameterGroupInfo(group_name="list")
                ]
            ),
            CommandParameter(
                name="start",
                type=ParameterType.Boolean,
                description="Flag to indicate asking launchd to start a service",
                default_value=True,
                parameter_group_info=[
                    ParameterGroupInfo(group_name="start")
                ]
            ),
            CommandParameter(
                name="stop",
                type=ParameterType.Boolean,
                description="Flag to indicate asking launchd to stop a service",
                default_value=True,
                parameter_group_info=[
                    ParameterGroupInfo(group_name="stop")
                ]
            ),
            CommandParameter(
                name="program",
                type=ParameterType.String,
                description="Program/binary to execute",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="submit")
                ]
            ),
            CommandParameter(
                name="load",
                display_name="Flag to indicate the load command",
                type=ParameterType.Boolean,
                default_value=True,
                description="Must be True to run 'load' with a file path to a plist file",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="load")
                ]
            ),
            CommandParameter(
                name="unload",
                display_name="Flag to indicate the unload command",
                type=ParameterType.Boolean,
                default_value=True,
                description="Must be True to run 'unload' with a file path to a plist file",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="unload")
                ]
            ),
            CommandParameter(
                name="file",
                display_name="Path to file to load/unload on target",
                type=ParameterType.String,
                description="Path to the plist file on disk to load/unload",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="unload"),
                    ParameterGroupInfo(group_name="load")
                ]
            ),
            CommandParameter(
                name="servicename",
                type=ParameterType.String,
                description="Name of the service to communicate with. Used with the submit, send, start/stop commands",
                default_value="",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="send"),
                    ParameterGroupInfo(group_name="submit"),
                    ParameterGroupInfo(group_name="start"),
                    ParameterGroupInfo(group_name="stop"),
                    ParameterGroupInfo(group_name="status"),
                ]
            ),
            CommandParameter(
                name="pid",
                type=ParameterType.Number,
                description="PID of the process",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="procinfo")
                ]
            ),
            CommandParameter(
                name="data",
                type=ParameterType.String,
                description="base64 encoded json data to send to a target service",
                parameter_group_info=[
                    ParameterGroupInfo(group_name="send")
                ]
            ),
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)
        await self.adjust_values()

    async def parse_dictionary(self, dictionary):
        self.load_args_from_dictionary(dictionary)
        await self.adjust_values()

    async def adjust_values(self):
        groupName = self.get_parameter_group_name()
        # python3.8 doesn't have switch statements, so have to use if statements
        self.add_arg("command", groupName, parameter_group_info=[ParameterGroupInfo(group_name=groupName)])
        self.remove_arg(groupName)


class XpcCommand(CommandBase):
    cmd = "xpc"
    needs_admin = False
    help_cmd = "xpc"
    description = "Use xpc to execute routines with launchd or communicate with another service/process."
    version = 1
    author = "@xorrior"
    argument_class = XpcArguments
    attributes = CommandAttributes(
        # uncomment when poseidon can dynamically compile commands
        supported_os=[SupportedOS.MacOS]
    )
    attackmapping = ["T1559"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
