#include <iostream>
#include <wx/wxprec.h>
#include <wx/filedlg.h>
#include <fstream>
#include <map>
#include <wx/dataview.h>
#include "rapidxml.hpp"
#include "base64.h"
#ifndef WX_PRECOMP
#include <wx/wx.h>
#endif

std::string string_to_hex(const std::string &input)
{
    static const char hex_digits[] = "0123456789ABCDEF";

    std::string output;
    output.reserve(input.length() * 2);
    for (unsigned char c : input)
    {
        output.push_back(hex_digits[c >> 4]);
        output.push_back(hex_digits[c & 15]);
    }
    return output;
}

std::string &ltrim(std::string &s)
{
    auto it = std::find_if(s.begin(), s.end(),
                           [](char c)
                           {
                               return !std::isspace<char>(c, std::locale::classic());
                           });
    s.erase(s.begin(), it);
    return s;
}

std::string &rtrim(std::string &s)
{
    auto it = std::find_if(s.rbegin(), s.rend(),
                           [](char c)
                           {
                               return !std::isspace<char>(c, std::locale::classic());
                           });
    s.erase(it.base(), s.end());
    return s;
}

std::string &trim(std::string &s)
{
    return ltrim(rtrim(s));
}

rapidxml::xml_document<> doc;
rapidxml::xml_node<> *root_node;

class Node;
using NodePtr = std::unique_ptr<Node>;
using NodePtrArray = std::vector<NodePtr>;

std::string getDisplayType(char *data)
{
    std::string str(data);
    if (str == "dict")
        return "Dictionary";
    if (str == "array")
        return "Array";
    if (str == "string")
        return "String";
    if (str == "integer" || str == "real")
        return "Number";
    if (str == "data")
        return "Data";
    if (str == "false" || str == "true")
        return "Boolean";
    if (str == "date")
        return "Date";
    return str;
}

class Node
{
public:
    Node(Node *parent,
         const wxString &key, const wxString &type,
         const wxString &value)
    {
        m_parent = parent;
        m_key = key;
        m_type = type;
        m_value = value;
    }

    ~Node() = default;

    bool IsContainer() const
    {
        return m_children.size() > 0;
    }

    Node *GetParent()
    {
        return m_parent;
    }
    NodePtrArray &GetChildren()
    {
        return m_children;
    }
    Node *GetNthChild(unsigned int n)
    {
        return m_children.at(n).get();
    }
    Node *AddEntry(const wxString &key, const wxString &type,
                   const wxString &value)
    {
        Node *node = new Node(this, key, type, value);
        m_children.push_back(NodePtr(node));
        m_value = std::to_string(m_children.size()) + " children";
        return node;
    }
    void Append(Node *child)
    {
        m_children.push_back(NodePtr(child));
        m_value = std::to_string(m_children.size()) + " children";
    }
    unsigned int GetChildCount() const
    {
        return m_children.size();
    }

public:
    wxString m_key;
    wxString m_type;
    wxString m_value;

private:
    Node *m_parent;
    NodePtrArray m_children;
};

class Model : public wxDataViewModel
{
public:
    Model();
    ~Model()
    {
        delete m_root;
    }
    virtual void GetValue(wxVariant &variant,
                          const wxDataViewItem &item, unsigned int col) const override;
    virtual bool SetValue(const wxVariant &variant,
                          const wxDataViewItem &item, unsigned int col) override;
    virtual wxDataViewItem GetParent(const wxDataViewItem &item) const override;
    virtual bool IsContainer(const wxDataViewItem &item) const override;
    virtual wxDataViewItem GetRoot() const;
    virtual void DeleteAll();
    virtual void SetPlistType(const std::string &type);
    virtual unsigned int GetChildren(const wxDataViewItem &parent,
                                     wxDataViewItemArray &array) const override;
    Node *AddRootEntry(const wxString &key, const wxString &type,
                       const wxString &value) const;
    virtual bool HasContainerColumns(const wxDataViewItem &item) const override;

private:
    Node *m_root;
};

Model::Model()
{
    m_root = new Node(nullptr, "Root", "Dictionary", "0 children");
};
Node *Model::AddRootEntry(const wxString &key, const wxString &type,
                          const wxString &value) const
{
    Node *node = new Node(m_root, key, type, value);
    m_root->Append(node);
    return node;
};

bool Model::HasContainerColumns(const wxDataViewItem &item) const
{
    return true;
}

void Model::GetValue(wxVariant &variant,
                     const wxDataViewItem &item, unsigned int col) const
{
    wxASSERT(item.IsOk());

    Node *node = (Node *)item.GetID();
    switch (col)
    {
    case 0:
        variant = node->m_key;
        break;
    case 1:
        variant = node->m_type;
        break;
    case 2:
        variant = node->m_value;
        break;
    default:
        wxLogError("Model::GetValue: wrong column %d", col);
    }
};

wxDataViewItem Model::GetRoot() const
{
    return wxDataViewItem((void *)m_root);
}

bool Model::SetValue(const wxVariant &variant,
                     const wxDataViewItem &item, unsigned int col)
{
    wxASSERT(item.IsOk());

    Node *node = (Node *)item.GetID();
    switch (col)
    {
    case 0:
        node->m_key = variant.GetString();
        return true;
    case 1:
        node->m_type = variant.GetString();
        return true;
    case 2:
        node->m_value = variant.GetString();
        return true;

    default:
        wxLogError("Model::SetValue: wrong column");
    }
    return false;
}

void Model::DeleteAll()
{
    m_root = new Node(nullptr, "Root", "Dictionary", "0 children");
}
void Model::SetPlistType(const std::string &type)
{
    m_root->m_type = type;
}

wxDataViewItem Model::GetParent(const wxDataViewItem &item) const
{
    if (!item.IsOk())
        return wxDataViewItem(0);

    Node *node = (Node *)item.GetID();

    if (node == m_root)
        return wxDataViewItem(0);

    return wxDataViewItem((void *)node->GetParent());
}

bool Model::IsContainer(const wxDataViewItem &item) const
{
    if (!item.IsOk())
        return true;

    Node *node = (Node *)item.GetID();
    return node->IsContainer();
}

unsigned int Model::GetChildren(const wxDataViewItem &parent,
                                wxDataViewItemArray &array) const
{
    Node *node = (Node *)parent.GetID();
    if (!node)
    {
        array.Add(wxDataViewItem((void *)m_root));
        return 1;
    }

    if (node->GetChildCount() == 0)
    {
        return 0;
    }

    for (const auto &child : node->GetChildren())
    {
        array.Add(wxDataViewItem(child.get()));
    }

    return array.size();
}

class MyApp : public wxApp
{
public:
    virtual bool OnInit();
};
class Frame : public wxFrame
{
public:
    Frame(const wxString &title, const wxPoint &pos, const wxSize &size);

private:
    void OnFileOpen(wxCommandEvent &event);
    void OnExit(wxCommandEvent &event);
    void OnAbout(wxCommandEvent &event);
    wxDataViewCtrl *dataview;
    Model *model;
    wxDECLARE_EVENT_TABLE();
};

enum
{
    ID_FILE = 1,
    ID_NEW = 2
};

wxBEGIN_EVENT_TABLE(Frame, wxFrame)
    EVT_MENU(ID_FILE, Frame::OnFileOpen)
        EVT_MENU(wxID_EXIT, Frame::OnExit)
            EVT_MENU(wxID_ABOUT, Frame::OnAbout)
                wxEND_EVENT_TABLE()
                    wxIMPLEMENT_APP(MyApp);

std::string dataString(std::string data)
{
    std::string str = "";
    for (int i = 0; i < data.length(); i++)
    {
        str += data[i];
        if ((i + 1) % 8 == 0 && i != data.length() - 1)
        {
            str += " ";
        }
    }
    return str;
}

Node *AddNodes(Node *treeNode, rapidxml::xml_node<char> *node, std::string k)
{
    int index = 0;
    if (k == "array")
    {
        for (rapidxml::xml_node<> *n = node->first_node(); n; n = n->next_sibling())
        {
            std::string key = std::string(n->name());
            std::string value = n->value();
            value = trim(value);

            Node *tn = treeNode->AddEntry(std::to_string(index), getDisplayType(n->name()), value);
            if (std::string(n->name()) == "true")
            {
                tn->m_value = "True";
            }
            else if (std::string(n->name()) == "false")
            {
                tn->m_value = "False";
            }
            tn = AddNodes(tn, n, key);

            index += 1;
        }
    }
    else
    {
        for (rapidxml::xml_node<> *n = node->first_node(); n; n = n->next_sibling())
        {
            std::string key = std::string(n->name());
            if (key != "key")
            {
                continue;
            }
            else
            {
                std::string value = std::string(n->next_sibling()->value());
                if (std::string(n->next_sibling()->name()) == "data")
                {
                    value = dataString(string_to_hex(base64::from_base64(trim(value))));
                }
                Node *tn = treeNode->AddEntry(n->value(), getDisplayType(n->next_sibling()->name()), value);
                if (std::string(n->next_sibling()->name()) == "true")
                {
                    tn->m_value = "True";
                }
                else if (std::string(n->next_sibling()->name()) == "false")
                {
                    tn->m_value = "False";
                }
                tn = AddNodes(tn, n->next_sibling(), n->next_sibling()->name());
            }
        }
    }
    return treeNode;
}

bool MyApp::OnInit()
{
    Frame *frame = new Frame("Qlist", wxPoint(50, 50), wxSize(750, 650));
    frame->Show(true);
    return true;
}
Frame::Frame(const wxString &title, const wxPoint &pos, const wxSize &size)
    : wxFrame(NULL, wxID_ANY, title, pos, size)
{
    wxMenu *menuFile = new wxMenu;
    menuFile->Append(ID_FILE, "&Open\tCtrl-O");
    menuFile->Append(ID_NEW, "&New\tCtrl-N");
    menuFile->Append(wxID_EXIT);
    menuFile->Append(wxID_ABOUT, "&About Qlist");
    menuFile->Append(wxID_PREFERENCES, "&Settings");
    wxMenuBar *menuBar = new wxMenuBar;
    menuBar->Append(menuFile, "&File");
    SetMenuBar(menuBar);
    dataview = new wxDataViewCtrl(this, wxID_ANY, wxDefaultPosition, wxDefaultSize, wxDV_ROW_LINES);
    dataview->AppendTextColumn("Key", 0, wxDATAVIEW_CELL_EDITABLE, (40.0 / 100.0) * size.GetWidth());
    wxArrayString *choices = new wxArrayString();
    choices->Add("Array");
    choices->Add("Dictionary");
    choices->Add("String");
    choices->Add("Number");
    choices->Add("Data");
    choices->Add("Date");
    choices->Add("Boolean");
    wxDataViewChoiceRenderer *choice = new wxDataViewChoiceRenderer(*choices);
    wxDataViewColumn *typeColumn = new wxDataViewColumn("Type", choice, 1);
    dataview->AppendColumn(typeColumn);
    dataview->AppendTextColumn("Value", 2);
    model = new Model();
}
void Frame::OnExit(wxCommandEvent &event)
{
    Close(true);
}
void Frame::OnAbout(wxCommandEvent &event)
{
    wxMessageBox("This is a wxWidgets' Hello world sample",
                 "About Hello World", wxOK | wxICON_INFORMATION);
}
void Frame::OnFileOpen(wxCommandEvent &event)
{
    model->DeleteAll();
    wxFileDialog openFileDialog(this, _("Open Property-List file"), wxEmptyString, wxEmptyString, _("Property-List file|*.plist"), wxFD_OPEN | wxFD_FILE_MUST_EXIST);
    if (openFileDialog.ShowModal() == wxID_OK)
    {

        std::ifstream PlistFile(openFileDialog.GetPath());
        std::vector<char> buffer((std::istreambuf_iterator<char>(PlistFile)), std::istreambuf_iterator<char>());
        buffer.push_back('\0');
        doc.parse<0>(&buffer[0]);
        int index = 0;
        rapidxml::xml_node<char> *root = doc.first_node("plist")->first_node();
        for (rapidxml::xml_node<> *node = root->first_node(); node; node = node->next_sibling())
        {
            std::string key = std::string(node->name());
            if (std::string(root->name()) == "array")
            {
                Node *treeNode = model->AddRootEntry(std::to_string(index), getDisplayType(node->name()), node->value());
                treeNode = AddNodes(treeNode, node, std::string(node->name()));
                model->SetPlistType("Array");
            }
            else
            {
                if (key != "key")
                {
                    continue;
                }
                Node *treeNode = model->AddRootEntry(node->value(), getDisplayType(node->next_sibling()->name()), node->next_sibling()->value());
                treeNode = AddNodes(treeNode, node->next_sibling(), std::string(node->next_sibling()->name()));
            }
            index += 1;
        }
        dataview->AssociateModel(model);
        dataview->ExpandChildren(model->GetRoot());
    }
    else
    {
        return;
    }
}